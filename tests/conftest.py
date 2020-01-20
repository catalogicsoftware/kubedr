
import configparser
import json
import os
import pprint

import pytest

from src.common import kubeclient

env_config_data = [
    ("wait_for_res_to_appear_num_attempts", 15, int),
    ("wait_for_res_to_appear_interval_secs", 1, int),
    ("wait_for_pod_to_be_done_num_attempts", 100, int),
    ("wait_for_pod_to_be_done_interval_secs", 3, int)
]

# This class encapsulates all the parameters that can be controlled
# using env variables.
class EnvConfig:
    def __init__(self):
        for name, default_val, factory in env_config_data:
            self._set_env_config(name, default_val, factory)

    def _set_env_config(self, name, default_val, factory):
        setattr(self, name, factory(os.environ.get(name.upper(), default_val)))

class GlobalConfig:
    def __init__(self, envconfig):
        self.envconfig = envconfig
        self.restic_password = "testpass"

        self.rootdir = os.environ['TESTS_ROOTDIR']

        iniconfig = configparser.ConfigParser()
        # pytest has a way of finding the path of "pytest.ini" using "config"
        # object but it is not very well documented. So for now, directly 
        # construct the path.
        iniconfig.read(os.path.join(self.rootdir, "pytest.ini"))
        self.iniconfig = iniconfig

        self.configdir = os.path.join(self.rootdir, "config")

        self.testenv = None
        testenv_f = os.path.join(self.configdir, "testenv.json")
        if os.path.exists(testenv_f):
            self.testenv = json.load(open(testenv_f))

        self._init_apis()

    def _init_apis(self):
        self.namespace = "kubedr-system"
        self.pod_api = kubeclient.PodAPI(self.namespace)
        self.backuploc_api = kubeclient.BackupLocationAPI(self.namespace)
        self.mbp_api = kubeclient.MetadataBackupPolicyAPI(self.namespace)
        self.secret_api = kubeclient.SecretAPI(self.namespace)

# This is being set as a global variable so that library code
# such as "kubeclient" can easily access the configuration set
# through env variables.
envconfig = EnvConfig()

@pytest.fixture(scope = "session")
def globalconfig():
    kubeclient.init()
    pprint.pprint(envconfig.__dict__)
    return GlobalConfig(envconfig)

