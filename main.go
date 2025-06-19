package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Global variable for the database connection
var db *sql.DB

// Game represents a single game record
type Game struct {
	ID      int       `json:"id"`
	Date    time.Time `json:"date"`
	Players []Player  `json:"players"`
}

// Player represents a player and their score in a game
type Player struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// initDB initializes the SQLite database and creates the necessary tables
// if they don't already exist.
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./everdell.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	createGamesTableSQL := `
	CREATE TABLE IF NOT EXISTS games (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"date" DATETIME
	);`
	statement, err := db.Prepare(createGamesTableSQL)
	if err != nil {
		log.Fatal("Failed to prepare games table statement:", err)
	}
	statement.Exec()

	createPlayersTableSQL := `
	CREATE TABLE IF NOT EXISTS players (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"game_id" INTEGER,
		"name" TEXT,
		"score" INTEGER,
		FOREIGN KEY(game_id) REFERENCES games(id)
	);`
	statement, err = db.Prepare(createPlayersTableSQL)
	if err != nil {
		log.Fatal("Failed to prepare players table statement:", err)
	}
	statement.Exec()
	log.Println("Database initialized successfully")
}

// gamesHandler routes requests based on the HTTP method.
func gamesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getGames(w, r)
	case "POST":
		addGame(w, r)
	case "PUT":
		updateGame(w,r)
	case "DELETE":
		deleteGame(w,r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getGames retrieves all games from the database and returns them as JSON.
func getGames(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, date FROM games ORDER BY date DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var dateStr string
		if err := rows.Scan(&g.ID, &dateStr); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
        g.Date, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", dateStr)
        if err != nil {
             g.Date, err = time.Parse(time.RFC3339, dateStr)
             if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
             }
        }

		playerRows, err := db.Query("SELECT name, score FROM players WHERE game_id = ?", g.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer playerRows.Close()

		for playerRows.Next() {
			var p Player
			if err := playerRows.Scan(&p.Name, &p.Score); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			g.Players = append(g.Players, p)
		}
		games = append(games, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

// parseGameData is a helper function to parse game data from a form.
func parseGameData(r *http.Request) (time.Time, []string, []int, error) {
    if err := r.ParseForm(); err != nil {
        return time.Time{}, nil, nil, fmt.Errorf("failed to parse form: %v", err)
    }

    gameDateStr := r.FormValue("date")
    playerNamesStr := r.FormValue("playerNames")
    playerScoresStr := r.FormValue("playerScores")
    
    playerNames := strings.Split(playerNamesStr, ",")
    playerScoresStrArr := strings.Split(playerScoresStr, ",")
    
    var playerScores []int
    for _, s := range playerScoresStrArr {
		if s == "" { continue }
        score, err := strconv.Atoi(s)
        if err != nil {
            return time.Time{}, nil, nil, fmt.Errorf("invalid score value: %v", err)
        }
        playerScores = append(playerScores, score)
    }

	if len(playerNames) != len(playerScores) {
		return time.Time{}, nil, nil, fmt.Errorf("mismatched number of players and scores")
	}
    
    var gameDate time.Time
    var err error
    if gameDateStr == "" {
        gameDate = time.Now()
    } else {
        gameDate, err = time.Parse("2006-01-02", gameDateStr)
        if err != nil {
            return time.Time{}, nil, nil, fmt.Errorf("invalid date format. Please use YYYY-MM-DD")
        }
    }
	return gameDate, playerNames, playerScores, nil
}


// addGame adds a new game record to the database.
func addGame(w http.ResponseWriter, r *http.Request) {
    gameDate, playerNames, playerScores, err := parseGameData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec("INSERT INTO games (date) VALUES (?)", gameDate)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := 0; i < len(playerNames); i++ {
		_, err := tx.Exec("INSERT INTO players (game_id, name, score) VALUES (?, ?, ?)", gameID, playerNames[i], playerScores[i])
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()
	w.WriteHeader(http.StatusCreated)
}

// updateGame modifies an existing game record in the database.
func updateGame(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("id")
	if gameIDStr == "" {
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	gameDate, playerNames, playerScores, err := parseGameData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the game date
	_, err = tx.Exec("UPDATE games SET date = ? WHERE id = ?", gameDate, gameID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update game date: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete old players for this game
	_, err = tx.Exec("DELETE FROM players WHERE game_id = ?", gameID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete old players: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert new players for this game
	for i := 0; i < len(playerNames); i++ {
		_, err := tx.Exec("INSERT INTO players (game_id, name, score) VALUES (?, ?, ?)", gameID, playerNames[i], playerScores[i])
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to insert new players: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()
	w.WriteHeader(http.StatusOK)
}


// deleteGame deletes a game record from the database.
func deleteGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM players WHERE game_id = ?", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec("DELETE FROM games WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        tx.Rollback()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if rowsAffected == 0 {
        tx.Rollback()
        http.Error(w, "Game not found", http.StatusNotFound)
        return
    }
	tx.Commit()
	w.WriteHeader(http.StatusOK)
}


// indexHTML is the HTML content for the frontend.
const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Everdell Game Tracker</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', sans-serif; }
		.form-input {
			@apply mt-1 block w-full border-gray-300 shadow-sm focus:border-green-500 focus:ring-green-500 text-sm px-3 py-2;
		}
		.btn {
			@apply py-2 px-4 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-offset-2;
		}
		.btn-primary {
			@apply bg-green-600 text-white hover:bg-green-700 focus:ring-green-500;
		}
		.btn-secondary {
			@apply bg-gray-200 text-gray-700 hover:bg-gray-300 focus:ring-gray-400;
		}
		.btn-sm {
			@apply py-1 px-2 text-xs;
		}
    </style>
</head>
<body class="bg-gray-100 text-gray-800">

    <div class="container mx-auto p-4 md:p-6 max-w-2xl">
        <header class="text-center mb-6">
            <h1 class="text-3xl font-bold text-green-800">Everdell Game Tracker</h1>
            <p class="text-md text-gray-600">Keep a record of your woodland adventures!</p>
        </header>

        <main>
            <!-- History View -->
            <div id="historyView">
                <div class="flex justify-between items-center mb-4">
                    <h2 class="text-xl font-semibold">Game History</h2>
                    <button id="showAddFormBtn" class="btn btn-primary">Add New Game</button>
                </div>
                <div class="bg-white p-4 shadow">
                    <div id="gameList" class="space-y-3">
                        <!-- Game records will be inserted here -->
                    </div>
                </div>
            </div>

            <!-- Form View (hidden by default) -->
            <div id="formView" class="hidden">
                <div class="bg-white p-4 shadow">
                    <h2 id="formTitle" class="text-xl font-semibold mb-3 border-b pb-2">Add New Game</h2>
                    <form id="gameForm">
						<input type="hidden" id="gameId" name="gameId">
                        <div class="mb-3">
                            <label for="date" class="block text-sm font-medium text-gray-700">Date</label>
                            <input type="date" id="date" name="date" class="form-input">
                        </div>

                        <div id="playersContainer">
                            <!-- Player entries will be added here -->
                        </div>

                        <div class="flex justify-between items-center my-3">
                            <button type="button" id="addPlayerBtn" class="text-sm text-green-600 hover:text-green-800 font-medium">+ Add Player</button>
                        </div>

						<div class="flex justify-end gap-2 mt-4 border-t pt-4">
							<button type="button" id="cancelBtn" class="btn btn-secondary">Cancel</button>
							<button type="submit" class="btn btn-primary">Save Game</button>
						</div>
                    </form>
                </div>
            </div>
        </main>
    </div>

    <script>
        document.addEventListener('DOMContentLoaded', () => {
			// Views and major elements
            const historyView = document.getElementById('historyView');
            const formView = document.getElementById('formView');
            const gameList = document.getElementById('gameList');
			
			// Form elements
            const gameForm = document.getElementById('gameForm');
            const formTitle = document.getElementById('formTitle');
			const gameIdInput = document.getElementById('gameId');
            const dateInput = document.getElementById('date');
            const playersContainer = document.getElementById('playersContainer');
            
			// Buttons
			const showAddFormBtn = document.getElementById('showAddFormBtn');
            const cancelBtn = document.getElementById('cancelBtn');
            const addPlayerBtn = document.getElementById('addPlayerBtn');
			
			// App state
			let gamesCache = [];

            // --- View Switching ---
			function showFormView() { 
				historyView.classList.add('hidden');
                formView.classList.remove('hidden');
			}
			function showHistoryView() {
				historyView.classList.remove('hidden');
                formView.classList.add('hidden');
			}

            showAddFormBtn.addEventListener('click', () => {
				formTitle.textContent = "Add New Game";
                resetAndSetupForm();
				showFormView();
            });

            cancelBtn.addEventListener('click', () => {
                showHistoryView();
            });

            // --- Form Logic ---
            function resetAndSetupForm() {
                gameForm.reset();
				gameIdInput.value = '';
                playersContainer.innerHTML = '';
                addPlayerField(); // Add first player
                addPlayerField(); // Add second player
            }

            function addPlayerField(name = '', score = '') {
                const playerCount = playersContainer.children.length + 1;
                const newPlayerEntry = document.createElement('div');
                newPlayerEntry.classList.add('player-entry', 'mb-2', 'p-2', 'border', 'relative');
                
                let removeBtnHtml = '';
                if (playerCount > 1) { // Allow removing all but the first player
                     removeBtnHtml = '<button type="button" class="remove-player-btn absolute top-1 right-1 text-red-500 hover:text-red-700 text-xs font-bold">X</button>';
                }

                newPlayerEntry.innerHTML = 
                    '<div class="flex justify-between items-center">' +
                    '    <label class="block text-sm font-medium text-gray-700">Player ' + playerCount + '</label>' +
                         removeBtnHtml +
                    '</div>' +
                    '<input type="text" name="playerName" placeholder="Player Name" class="form-input" value="' + name + '" required>' +
                    '<input type="number" name="playerScore" placeholder="Score" class="form-input mt-2" value="' + score + '" required>';
                
                playersContainer.appendChild(newPlayerEntry);
                attachRemoveListeners();
            }
            
            addPlayerBtn.addEventListener('click', () => addPlayerField());

            function updatePlayerLabels() {
                document.querySelectorAll('.player-entry').forEach((entry, index) => {
                    entry.querySelector('label').textContent = 'Player ' + (index + 1);
                });
            }

            function attachRemoveListeners() {
                 document.querySelectorAll('.remove-player-btn').forEach(button => {
                    // Replace the button to remove old listeners before adding a new one
                    const newButton = button.cloneNode(true);
                    button.parentNode.replaceChild(newButton, button);
                    newButton.addEventListener('click', function() {
                        this.closest('.player-entry').remove();
                        updatePlayerLabels();
                    });
                });
            }

            gameForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const playerEntries = document.querySelectorAll('.player-entry');
                let playerNames = [];
                let playerScores = [];

                playerEntries.forEach(entry => {
                    const nameInput = entry.querySelector('input[name="playerName"]');
                    const scoreInput = entry.querySelector('input[name="playerScore"]');
                    if (nameInput.value.trim() !== '') {
                        playerNames.push(nameInput.value.trim());
                        playerScores.push(parseInt(scoreInput.value, 10) || 0);
                    }
                });

                if (playerNames.length === 0) {
                    alert("Please enter at least one player.");
                    return;
                }
                
                const gameId = gameIdInput.value;
                const method = gameId ? 'PUT' : 'POST';
                const url = gameId ? '/games?id=' + gameId : '/games';

                const formData = new URLSearchParams();
                formData.append('date', dateInput.value);
                formData.append('playerNames', playerNames.join(','));
                formData.append('playerScores', playerScores.join(','));

                try {
                    const response = await fetch(url, { method: method, body: formData });
                    if (response.ok) {
                        showHistoryView();
                        fetchGames();
                    } else {
                        alert('Failed to save game: ' + await response.text());
                    }
                } catch (error) {
                    alert('An error occurred while saving the game.');
                }
            });

            // --- Game History Logic ---
            async function fetchGames() {
                try {
                    const response = await fetch('/games');
                    gamesCache = await response.json();
                    renderGameList();
                } catch (error) {
                    gameList.innerHTML = '<p class="text-red-500">Could not load game history.</p>';
                }
            }

			function renderGameList() {
				gameList.innerHTML = ''; 
				if (!gamesCache || gamesCache.length === 0) {
					gameList.innerHTML = '<p class="text-gray-500 p-2">No games recorded yet.</p>';
					return;
				}

				gamesCache.forEach(game => {
					const gameElement = document.createElement('div');
					gameElement.classList.add('p-3', 'border', 'bg-gray-50', 'relative');
					const gameDate = new Date(game.date);
					// Format date as YYYY-MM-DD for the input field
					const dateForInput = gameDate.toISOString().split('T')[0];
					const displayDate = gameDate.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
					
					game.players.sort((a, b) => b.score - a.score);

					let playersHtml = '<ul class="list-disc list-inside mt-2 text-sm">';
					game.players.forEach(player => {
						playersHtml += '<li class="text-gray-700">' + player.name + ': <span class="font-semibold">' + player.score + '</span></li>';
					});
					playersHtml += '</ul>';
					
					const buttonsHtml = 
						'<div class="absolute top-2 right-2 flex gap-2">' +
						'<button class="edit-game-btn btn btn-sm btn-secondary" data-game-id="' + game.id + '">Edit</button>' +
						'<button class="delete-game-btn btn btn-sm btn-secondary" data-game-id="' + game.id + '">X</button>' +
						'</div>';

					gameElement.innerHTML = buttonsHtml + '<h3 class="font-semibold text-md text-green-700">Game on ' + displayDate + '</h3>' + playersHtml;
					gameList.appendChild(gameElement);
				});
				attachHistoryButtonsListeners();
			}

			function attachHistoryButtonsListeners() {
				// Delete buttons
				document.querySelectorAll('.delete-game-btn').forEach(button => {
					button.addEventListener('click', async function() {
						const gameId = this.dataset.gameId;
						if (confirm('Are you sure you want to delete this game?')) {
							try {
								const response = await fetch('/games?id=' + gameId, { method: 'DELETE' });
								if (response.ok) fetchGames();
								else alert('Failed to delete game: ' + await response.text());
							} catch (error) {
								alert('An error occurred while deleting the game.');
							}
						}
					});
				});

				// Edit buttons
				document.querySelectorAll('.edit-game-btn').forEach(button => {
					button.addEventListener('click', function() {
						const gameId = parseInt(this.dataset.gameId, 10);
						const gameToEdit = gamesCache.find(g => g.id === gameId);
						if (gameToEdit) {
							formTitle.textContent = "Edit Game";
							gameIdInput.value = gameToEdit.id;
							dateInput.value = new Date(gameToEdit.date).toISOString().split('T')[0];
							
							playersContainer.innerHTML = '';
							gameToEdit.players.forEach(p => addPlayerField(p.name, p.score));

							showFormView();
						}
					});
				});
			}
            
            // Initial Load
            fetchGames();
        });
    </script>

</body>
</html>
`

// main is the entry point of the application.
func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, indexHTML)
	})
	http.HandleFunc("/games", gamesHandler)

	port := "9797"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.Printf("Server starting on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
