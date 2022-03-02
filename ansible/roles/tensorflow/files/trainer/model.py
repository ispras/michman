from tensorflow.keras.layers import concatenate, Input, Dense, Reshape
from tensorflow.keras import Model


def create_model(features, layer_sizes, loss, optimizer, metrics=None):
    inputs = [Input(shape=(1), name=feature) for feature in features]

    common = concatenate(inputs)
    for layer_size in layer_sizes:
        common = Dense(layer_size)(common)
    common = Dense(1)(common)
    output = Reshape(target_shape=(), name='target')(common)

    keras_model = Model(inputs=inputs, outputs=output)
    keras_model.compile(loss=loss, optimizer=optimizer, metrics=metrics)

    return keras_model
