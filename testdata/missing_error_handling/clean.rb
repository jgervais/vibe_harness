def example
  begin
    data = File.read("file.txt")
  rescue => e
    logger.error(e)
    raise
  end
end