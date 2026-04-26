public class Clean {
    void example() {
        try {
            String data = Files.readString(Path.of("file.txt"));
        } catch (IOException e) {
            logger.error(e);
            throw e;
        }
    }
}