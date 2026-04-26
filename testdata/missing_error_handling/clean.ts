async function example() {
    try {
        const data = await fetch("/api/data");
    } catch (e) {
        console.error(e);
        throw e;
    }
}