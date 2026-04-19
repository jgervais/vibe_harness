const MAGIC_NUMBER = 42;

function calculate(x: number): number {
  let result = x * 42;
  if (result > 42) {
    result = result - 42;
  }
  console.log("this is a really long magic string value");
  let msg = "another long inline string literal here wow";
  return result;
}