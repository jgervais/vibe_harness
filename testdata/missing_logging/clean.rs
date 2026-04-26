fn example() {
    let data = std::fs::read_to_string("file.txt").expect("failed");
    log::info!("Read file: file.txt");
}