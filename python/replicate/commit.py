import datetime
import hashlib
import json
import random
from typing import Optional, Dict, Any

from .hash import random_hash
from .metadata import rfc3339_datetime
from .storage import Storage


# fmt: off
try:
    import numpy as np
    has_numpy = True
except ImportError:
    has_numpy = False
try:
    import torch
    has_torch = True
except ImportError:
    has_torch = False
try:
    import tensorflow as tf
    has_tensorflow = True
except ImportError:
    has_tensorflow = False
# fmt: on


class CustomJSONEncoder(json.JSONEncoder):
    def default(self, obj):
        if has_numpy:
            if isinstance(obj, np.integer):
                return int(obj)
            elif isinstance(obj, np.floating):
                return float(obj)
            elif isinstance(obj, np.ndarray):
                return obj.tolist()
        if has_torch and isinstance(obj, torch.Tensor):
            return obj.detach().tolist()
        if has_tensorflow and isinstance(obj, tf.Tensor):
            return obj.numpy().tolist()
        print(type(obj))
        return json.JSONEncoder.default(self, obj)


class Commit(object):
    """
    A snapshot of a training job -- the working directory plus any metadata.
    """

    def __init__(
        self,
        experiment,  # can't type annotate due to circular import
        project_dir: str,
        created: datetime.datetime,
        metrics: Dict[str, Any],
    ):
        self.experiment = experiment
        self.project_dir = project_dir
        self.created = created
        self.metrics = metrics

        # TODO (bfirsh): content addressable id
        self.id = random_hash()

    def save(self, storage: Storage):
        storage.put_directory(self.get_path(), self.project_dir)
        storage.put(
            self.get_path() + "replicate-metadata.json",
            json.dumps(
                {
                    "id": self.id,
                    "created": rfc3339_datetime(self.created),
                    "experiment": self.experiment.get_metadata(),
                    "metrics": self.metrics,
                },
                indent=2,
                cls=CustomJSONEncoder,
            ),
        )

    def get_path(self) -> str:
        return "commits/{}/".format(self.id)
