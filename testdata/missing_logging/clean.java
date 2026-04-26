import java.nio.file.Files;

public class Clean {
    void example() {
        String data = Files.readString(Path.of("file.txt"));
        logger.info("Read file: file.txt");
    }
}