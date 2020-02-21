import logging
import os
import pprint
import tempfile
import time
import traceback

from common import kubeclient

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

def timestamp():
    return int(time.time())

def create_hostpath_pv():
    pv_api = kubeclient.PersistentVolumeAPI()
    pv_name = "{}-{}".format("pv", timestamp())
    pv_dir = tempfile.mkdtemp()

    pv_spec = {
        "accessModes": ["ReadWriteOnce"],
        "capacity": {
            "storage": "2Gi"
        },
        "hostPath": {
            "path": pv_dir
        },
        "persistentVolumeReclaimPolicy": "Delete",
        "storageClassName": "standard",
        "volumeMode": "Filesystem"
    }

    return pv_api.create(pv_name, pv_spec)

def create_pvc_for_pv(pv):
    pprint.pprint(pv)
    pvc_api = kubeclient.PersistentVolumeClaimAPI(namespace="kubedr-system")
    name = "{}-{}".format("pvc", timestamp())

    spec = {
        "accessModes": ["ReadWriteOnce"],
        "resources": {
            "requests": {
                "storage": pv.spec.capacity["storage"]
            }
        },
        "volumeMode": "Filesystem",
        "volumeName": pv.metadata.name
    }

    pvc = pvc_api.create(name, spec)

    # Wait till PVC is bound
    for i in range(30):
        time.sleep(1)

        pvc = pvc_api.get(name)
        if pvc.status.phase == "Bound":
            return pvc

    raise Exception("PVC {} did not change to 'Bound' status.".format(name))



