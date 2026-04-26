fn main() {
    let result = do_something();
    match result {
        Ok(val) => println!("{}", val),
        Err(e) => return Err(e),
    }
}