try:
    do_something()
except ValueError as e:
    logging.error(e)
except OSError as e:
    logging.error(e)