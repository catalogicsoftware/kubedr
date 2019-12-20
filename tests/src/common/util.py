import logging
import os
import tempfile
import traceback

def ignore_errors(func):
    try:
        func()
    except:
        logging.error(traceback.format_exc())

def ignore_errors_pred(predicate, func):
    try:
        if predicate:
            func()
    except:
        logging.error(traceback.format_exc())
