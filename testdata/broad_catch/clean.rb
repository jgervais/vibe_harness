begin
  do_something
rescue ArgumentError => e
  logger.error(e)
end