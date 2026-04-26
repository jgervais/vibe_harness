begin
  do_something
rescue => e
  logger.error(e)
  raise
end