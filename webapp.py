import streamlit as st
import numpy as np
import cv2
import wave
import os


# Function to convert a message into binary format
def msgtobinary(msg):
    if type(msg) == str:
        result = ''.join([format(ord(i), "08b") for i in msg])
    elif type(msg) == bytes or type(msg) == np.ndarray:
        result = [format(i, "08b") for i in msg]
    elif type(msg) == int or type(msg) == np.uint8:
        result = format(msg, "08b")
    else:
        raise TypeError("Input type is not supported in this function")
    return result


# Image Steganography Functions
def encode_img_data(img, data):
    if (len(data) == 0):
        raise ValueError('Data entered to be encoded is empty')

    no_of_bytes = (img.shape[0] * img.shape[1] * 3) // 8
    if len(data) > no_of_bytes:
        raise ValueError("Insufficient bytes Error, Need Bigger Image or give Less Data !!")

    data += '*^*^*'
    binary_data = msgtobinary(data)
    length_data = len(binary_data)

    index_data = 0

    for i in img:
        for pixel in i:
            r, g, b = msgtobinary(pixel)
            if index_data < length_data:
                pixel[0] = int(r[:-1] + binary_data[index_data], 2)
                index_data += 1
            if index_data < length_data:
                pixel[1] = int(g[:-1] + binary_data[index_data], 2)
                index_data += 1
            if index_data < length_data:
                pixel[2] = int(b[:-1] + binary_data[index_data], 2)
                index_data += 1
            if index_data >= length_data:
                break
    return img


def decode_img_data(img):
    data_binary = ""
    for i in img:
        for pixel in i:
            r, g, b = msgtobinary(pixel)
            data_binary += r[-1]
            data_binary += g[-1]
            data_binary += b[-1]
            total_bytes = [data_binary[i: i + 8] for i in range(0, len(data_binary), 8)]
            decoded_data = ""
            for byte in total_bytes:
                decoded_data += chr(int(byte, 2))
                if decoded_data[-5:] == "*^*^*":
                    return decoded_data[:-5]


# Audio Steganography Functions
def encode_aud_data(file_bytes, data):
    song = wave.open(file_bytes, mode='rb')

    nframes = song.getnframes()
    frames = song.readframes(nframes)
    frame_bytes = bytearray(list(frames))

    data += '*^*^*'
    result = []
    for c in data:
        bits = bin(ord(c))[2:].zfill(8)
        result.extend([int(b) for b in bits])

    j = 0
    for i in range(len(result)):
        res = bin(frame_bytes[j])[2:].zfill(8)
        frame_bytes[j] = (frame_bytes[j] & 254) | result[i]
        j += 1

    frame_modified = bytes(frame_bytes)
    return frame_modified, song


def decode_aud_data(file_bytes):
    song = wave.open(file_bytes, mode='rb')

    nframes = song.getnframes()
    frames = song.readframes(nframes)
    frame_bytes = bytearray(list(frames))

    extracted = ""
    for i in range(len(frame_bytes)):
        res = bin(frame_bytes[i])[2:].zfill(8)
        extracted += res[-1]

        all_bytes = [extracted[i: i + 8] for i in range(0, len(extracted), 8)]
        decoded_data = ""
        for byte in all_bytes:
            decoded_data += chr(int(byte, 2))
            if decoded_data[-5:] == "*^*^*":
                return decoded_data[:-5]


# Video Steganography Functions
def embed_video_data(frame, data):
    data += '*^*^*'
    binary_data = msgtobinary(data)
    length_data = len(binary_data)

    index_data = 0

    for i in frame:
        for pixel in i:
            r, g, b = msgtobinary(pixel)
            if index_data < length_data:
                pixel[0] = int(r[:-1] + binary_data[index_data], 2)
                index_data += 1
            if index_data < length_data:
                pixel[1] = int(g[:-1] + binary_data[index_data], 2)
                index_data += 1
            if index_data < length_data:
                pixel[2] = int(b[:-1] + binary_data[index_data], 2)
                index_data += 1
            if index_data >= length_data:
                break
    return frame


def extract_video_data(frame):
    data_binary = ""
    for i in frame:
        for pixel in i:
            r, g, b = msgtobinary(pixel)
            data_binary += r[-1]
            data_binary += g[-1]
            data_binary += b[-1]

            total_bytes = [data_binary[i: i + 8] for i in range(0, len(data_binary), 8)]
            decoded_data = ""
            for byte in total_bytes:
                decoded_data += chr(int(byte, 2))
                if decoded_data[-5:] == "*^*^*":
                    return decoded_data[:-5]


# Web App using Streamlit
st.title("Steganography Web Application")

st.sidebar.title("Choose Steganography Method")
choice = st.sidebar.selectbox(
    "Select an option",
    ("Text Steganography", "Image Steganography", "Audio Steganography", "Video Steganography")
)

if choice == "Image Steganography":
    st.header("Image Steganography")
    option = st.radio("Choose an option", ('Encode', 'Decode'))

    if option == 'Encode':
        image_file = st.file_uploader("Upload an image", type=["jpg", "png"])
        if image_file is not None:
            img = cv2.imdecode(np.frombuffer(image_file.read(), np.uint8), 1)
            secret_message = st.text_input("Enter the secret message")
            if st.button("Encode"):
                stego_img = encode_img_data(img, secret_message)
                cv2.imwrite('encoded_image.png', stego_img)
                st.image(stego_img, caption='Encoded Image', use_column_width=True)

    elif option == 'Decode':
        image_file = st.file_uploader("Upload a stego image", type=["jpg", "png"])
        if image_file is not None:
            img = cv2.imdecode(np.frombuffer(image_file.read(), np.uint8), 1)
            if st.button("Decode"):
                hidden_message = decode_img_data(img)
                st.success(f"Hidden message: {hidden_message}")

elif choice == "Audio Steganography":
    st.header("Audio Steganography")
    option = st.radio("Choose an option", ('Encode', 'Decode'))

    if option == 'Encode':
        audio_file = st.file_uploader("Upload an audio file", type=["wav"])
        if audio_file is not None:
            secret_message = st.text_input("Enter the secret message")
            if st.button("Encode"):
                frame_modified, song = encode_aud_data(audio_file, secret_message)
                with wave.open('stego_audio.wav', 'wb') as fd:
                    fd.setparams(song.getparams())
                    fd.writeframes(frame_modified)
                st.audio('stego_audio.wav', format='audio/wav')

    elif option == 'Decode':
        audio_file = st.file_uploader("Upload a stego audio file", type=["wav"])
        if audio_file is not None:
            if st.button("Decode"):
                hidden_message = decode_aud_data(audio_file)
                st.success(f"Hidden message: {hidden_message}")

elif choice == "Video Steganography":
    st.header("Video Steganography")
    option = st.radio("Choose an option", ('Encode', 'Decode'))

    if option == 'Encode':
        video_file = st.file_uploader("Upload a video", type=["mp4"])
        if video_file is not None:
            secret_message = st.text_input("Enter the secret message")
            if st.button("Encode"):
                cap = cv2.VideoCapture(video_file)
                ret, frame = cap.read()
                frame = embed_video_data(frame, secret_message)
                st.video(frame)

    elif option == 'Decode':
        video_file = st.file_uploader("Upload a stego video", type=["mp4"])
        if video_file is not None:
            if st.button("Decode"):
                cap = cv2.VideoCapture(video_file)
                ret, frame = cap.read()
                hidden_message = extract_video_data(frame)
                st.success(f"Hidden message: {hidden_message}")

