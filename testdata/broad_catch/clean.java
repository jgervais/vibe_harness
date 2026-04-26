public class Clean {
    void example() {
        try {
            doSomething();
        } catch (IOException e) {
            logger.error(e);
        } catch (SQLException e) {
            logger.error(e);
        }
    }
}