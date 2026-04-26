import os

def example():
    try:
        data = open("file.txt").read()
    except IOError as e:
        logging.error(e)
        raise