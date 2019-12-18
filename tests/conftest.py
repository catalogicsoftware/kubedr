
import configparser
import json
import os

import pytest

from src.common import kubeclient

class GlobalConfig:
    def __init__(self):
        self.rootdir = os.environ['TESTS_ROOTDIR']

        iniconfig = configparser.ConfigParser()
        # pytest has a way of finding the path of "pytest.ini" using "config"
        # object but it is not very well documented. So for now, directly 
        # construct the path.
        iniconfig.read(os.path.join(self.rootdir, "pytest.ini"))
        self.iniconfig = iniconfig

        self.configdir = os.path.join(self.rootdir, "config")

        self.testenv = None
        print("aaa 1")
        testenv_f = os.path.join(self.configdir, "testenv.json")
        print(testenv_f)
        if os.path.exists(testenv_f):
            print("aaa 2")
            self.testenv = json.load(open(testenv_f))

        self._init_apis()

    def _init_apis(self):
        self.namespace = "kubedr-system"
        self.pod_api = kubeclient.PodAPI(self.namespace)
        self.backuploc_api = kubeclient.BackupLocationAPI(self.namespace)
        self.secret_api = kubeclient.SecretAPI(self.namespace)

@pytest.fixture(scope = "session")
def globalconfig():
    kubeclient.init()

    return GlobalConfig()
