try:
    do_something()
except Exception as e:
    logging.error(e)

try:
    do_something_else()
except:
    pass