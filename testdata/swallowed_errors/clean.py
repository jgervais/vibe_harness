try:
    do_something()
except ValueError as e:
    logging.error(e)
    raise