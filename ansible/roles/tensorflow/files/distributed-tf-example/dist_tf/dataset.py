import numpy as np
import tensorflow as tf


def mnist_dataset(batch_size):
  (x_train, y_train), _ = tf.keras.datasets.mnist.load_data()
  # The `x` arrays are in uint8 and have values in the [0, 255] range.
  # You need to convert them to float32 with values in the [0, 1] range.
  x_train = x_train / np.float32(255)
  y_train = y_train.astype(np.int64)
  train_dataset = tf.data.Dataset.from_tensor_slices(
      (x_train, y_train)).shuffle(60000)
  return train_dataset


def mnist_dataset_creator(global_batch_size, input_context):
    batch_size = input_context.get_per_replica_batch_size(global_batch_size)
    dataset = mnist_dataset(batch_size)
    dataset = dataset.shard(input_context.num_input_pipelines,
                            input_context.input_pipeline_id)
    dataset = dataset.batch(batch_size)
    return dataset
