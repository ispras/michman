### Installation

```shell
pip install -r requirements.txt
```

Note that it contains `tensorflow` dependency which should be probably removed.

### Example startup

First host:

```shell
TF_CONFIG=$(cat tf_config_worker1.json) python train_mnist.py -b 32 -e 3
```

Second host:

```shell
TF_CONFIG=$(cat tf_config_worker2.json) python train_mnist.py -b 32 -e 3
```
