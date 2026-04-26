import java.nio.file.Files;

public class Violating {
    void example() {
        String data = Files.readString(Path.of("file.txt"));
    }
}