<script>
    import "../app.css"
</script>

<main>
    <a href="/">
        <h1>Everdell Game Log</h1>
    </a>
    
    <slot></slot>
</main>

<style>
    main {
        display: flex;
        flex-direction: column;
        max-width: 400px;
        margin: 0 auto;
    }
</style>