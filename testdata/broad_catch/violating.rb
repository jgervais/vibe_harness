begin
  do_something
rescue Exception => e
  logger.error(e)
end