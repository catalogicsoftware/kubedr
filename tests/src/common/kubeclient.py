
import base64
import logging
import time
import urllib3

from kubernetes import client
from kubernetes import config as k8sconfig

import conftest

class KubeResourceAPI:
    def __init__(self, namespace="default"):
        self.namespace = namespace
        self.v1api = client.CoreV1Api()
        self.batch_v1beta1_api = client.BatchV1beta1Api()
        self.cr_api = client.CustomObjectsApi()

    def create_metadata(self, name):
        metadata = client.V1ObjectMeta()
        metadata.name = name

        return metadata

class KubedrV1AlphaResource(KubeResourceAPI):
    def __init__(self, namespace="default"):
        super().__init__(namespace)
        self.group = "kubedr.catalogicsoftware.com"
        self.version = "v1alpha1"
        self.apiVersion = "{}/{}".format(self.group, self.version)
        self.res = {
            "apiVersion": self.apiVersion
        }

        # These must be set by subclasses.
        self.kind = ""
        self.plural = ""

    def create(self, name, spec):
        self.res["kind"] = self.kind
        self.res["metadata"] = {"name": name}
        self.res["spec"] = spec

        self.cr_api.create_namespaced_custom_object(
            group=self.group, version=self.version, namespace=self.namespace, plural=self.plural,
            body=self.res)

    def delete(self, name):
        self.cr_api.delete_namespaced_custom_object(
            group=self.group, version=self.version, namespace=self.namespace, plural=self.plural,
            name=name, body=client.V1DeleteOptions())

    def get(self, name):
        return self.cr_api.get_namespaced_custom_object(
            group=self.group, version=self.version, namespace=self.namespace, plural=self.plural,
            name=name)

class SecretAPI(KubeResourceAPI):
    def __init__(self, namespace="default"):
        super().__init__(namespace)

    def create(self, name, data):
        body = client.V1Secret()
        body.data = data

        body.metadata = self.create_metadata(name)
        body.metadata.namespace = self.namespace

        self.v1api.create_namespaced_secret(body.metadata.namespace, body)

    def delete(self, name):
        self.v1api.delete_namespaced_secret(name, self.namespace, 
                                            body=client.V1DeleteOptions())

class PodAPI(KubeResourceAPI):
    def __init__(self, namespace="default"):
        super().__init__(namespace)

    def list(self, label_selector="", timeout_seconds=30):
        return self.v1api.list_namespaced_pod(self.namespace, label_selector=label_selector, 
                                              timeout_seconds=timeout_seconds)

    def read(self, name):
        return self.v1api.read_namespaced_pod(name, self.namespace)

    def delete(self, name):
        self.v1api.delete_namespaced_pod(name, self.namespace, 
                                         body=client.V1DeleteOptions())

class CronJobAPI(KubeResourceAPI):
    def __init__(self, namespace="default"):
        super().__init__(namespace)

    def list(self, label_selector="", timeout_seconds=30):
        return self.batch_v1beta1_api.list_namespaced_cron_job(self.namespace, label_selector=label_selector,
                                                               timeout_seconds=timeout_seconds)

class BackupLocationAPI(KubedrV1AlphaResource):
    def __init__(self, namespace="default"):
        super().__init__(namespace)
        self.kind = "BackupLocation"
        self.plural = "backuplocations"

class MetadataBackupPolicyAPI(KubedrV1AlphaResource):
    def __init__(self, namespace="default"):
        super().__init__(namespace)
        self.kind = "MetadataBackupPolicy"
        self.plural = "metadatabackuppolicies"

def create_backuploc_creds(name, access_key, secret_key, restic_password):
    creds_data = {
        "access_key": base64.b64encode(access_key.encode("utf-8")).decode("utf-8"),
        "secret_key": base64.b64encode(secret_key.encode("utf-8")).decode("utf-8"),
        "restic_repo_password": base64.b64encode(restic_password.encode("utf-8")).decode("utf-8")
    }
    secret_api = SecretAPI(namespace="kubedr-system")
    secret_api.create(name, creds_data)

def create_etcd_creds(name, ca_crt, client_crt, client_key):
    creds_data = {
        "ca.crt": base64.b64encode(open(ca_crt, "rb").read()).decode("utf-8"),
        "client.crt": base64.b64encode(open(client_crt, "rb").read()).decode("utf-8"),
        "client.key": base64.b64encode(open(client_key, "rb").read()).decode("utf-8")
    }
    secret_api = SecretAPI(namespace="kubedr-system")
    secret_api.create(name, creds_data)

def wait_for_pod_to_appear(label_selector):
    num_attempts = conftest.envconfig.wait_for_res_to_appear_num_attempts
    interval_secs = conftest.envconfig.wait_for_res_to_appear_interval_secs

    pod_api = PodAPI(namespace="kubedr-system")

    for i in range(num_attempts):
        time.sleep(interval_secs)

        pods = pod_api.list(label_selector=label_selector)
        if len(pods.items) > 0:
            return pods

    raise Exception("Timed out waiting for pod with label: {}.".format(label_selector))

def wait_for_cronjob_to_appear(label_selector):
    num_attempts = conftest.envconfig.wait_for_res_to_appear_num_attempts
    interval_secs = conftest.envconfig.wait_for_res_to_appear_interval_secs

    cronjob_api = CronJobAPI(namespace="kubedr-system")

    for i in range(num_attempts):
        time.sleep(interval_secs)

        cronjobs = cronjob_api.list(label_selector=label_selector)
        if len(cronjobs.items) > 0:
            return cronjobs

    raise Exception("Timed out waiting for cronjob with label: {}.".format(label_selector))

def wait_for_pod_to_be_done(pod_name):
    num_attempts = conftest.envconfig.wait_for_pod_to_be_done_num_attempts
    interval_secs = conftest.envconfig.wait_for_pod_to_be_done_interval_secs

    pod_api = PodAPI(namespace="kubedr-system")

    for i in range(num_attempts):
        time.sleep(interval_secs)

        pod = pod_api.read(pod_name)
        if pod.status.phase in ["Succeeded", "Failed"]:
            return pod

    raise Exception("pod {} did not finish in time.".format(pod_name))

def init():
    k8sconfig.debug = True
    logging.getLogger("urllib3").setLevel(logging.DEBUG)
    k8sconfig.load_kube_config()


