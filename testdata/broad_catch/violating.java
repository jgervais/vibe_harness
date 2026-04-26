public class Violating {
    void example() {
        try {
            doSomething();
        } catch (Exception e) {
            logger.error(e);
        }
    }
}