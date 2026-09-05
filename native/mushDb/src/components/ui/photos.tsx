import React, { useState } from 'react';
import { StyleSheet, Text, View, Button, Image, Alert } from 'react-native';
import * as ImagePicker from 'expo-image-picker';

export default function PhotoPicker() {
    const [image, setImage] = useState<string | undefined>(undefined);

    // Function to open device gallery and select an image
    const pickImage = async () => {
        // Request media library permissions
        const { status } = await ImagePicker.requestMediaLibraryPermissionsAsync();

        if (status !== 'granted') {
            Alert.alert('Permission Denied', 'Camera roll permissions needed to select existing pictures');
            return;
        }

        // Launch the system image library
        let result = await ImagePicker.launchImageLibraryAsync({
            mediaTypes: ImagePicker.MediaTypeOptions.Images, // Restrict to images only
            allowsEditing: true,                             // Enable cropping/editing crop box
            aspect: [4, 3],                                  // Maintain a 4:3 aspect ratio
            quality: 1,                                      // Highest image quality (0 to 1)
        });

        if (!result.canceled) {
            setImage(result.assets[0].uri);                  // Save the image URI to state
        }
    };

    // Function to open the system camera and take a new photo
    const takePicture = async () => {
        // Request camera permissions
        const { status } = await ImagePicker.requestCameraPermissionsAsync();

        if (status !== 'granted') {
            Alert.alert('Permission Denied', 'Camera permissions required to take pictures');
            return;
        }

        // Launch the system camera UI
        let result = await ImagePicker.launchCameraAsync({
            mediaTypes: ImagePicker.MediaTypeOptions.Images,
            allowsEditing: true,
            aspect: [4, 3],
            quality: 1,
        });

        if (!result.canceled) {
            setImage(result.assets[0].uri);                  // Save the image URI to state
        }
    };

    return (
        <View style={styles.container}>
            <Button title="Pick an image from camera roll" onPress={pickImage} />
            <View style={styles.spacer} />
            <Button title="Take a photo with camera" onPress={takePicture} />

            {image && <Image source={{ uri: image }} style={styles.image} />}
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: '#fff',
    },
    spacer: {
        height: 20,
    },
    image: {
        width: 200,
        height: 200,
        marginTop: 20,
        borderRadius: 10,
    },
});
