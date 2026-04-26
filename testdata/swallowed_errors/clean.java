public class Clean {
    void example() {
        try {
            doSomething();
        } catch (IOException e) {
            logger.error(e);
            throw e;
        }
    }
}