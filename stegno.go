package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"gocv.io/x/gocv"
	"strconv"
)

var ZWC = map[string]string{
	"00": "\u200C", "01": "\u202C", "11": "\u202D", "10": "\u200E",
}

func txtEncode(text string) {
	l := len(text)
	add := ""

	for i := 0; i < l; i++ {
		t := int(text[i])
		if t >= 32 && t <= 64 {
			t1 := t + 48
			t2 := t1 ^ 170 // 170: 10101010
			res := fmt.Sprintf("%08b", t2)
			add += "0011" + res
		} else {
			t1 := t - 48
			t2 := t1 ^ 170
			res := fmt.Sprintf("%08b", t2)
			add += "0110" + res
		}
	}
	res1 := add + "111111111111"
	fmt.Println("The string after binary conversion applying all the transformation :- " + res1)
	length := len(res1)
	fmt.Println("Length of binary after conversion:- ", length)

	fileContent, err := ioutil.ReadFile("text_test.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	word := strings.Fields(string(fileContent))

	i := 0
	for i < len(res1) {
		s := word[i/12]
		HM_SK := ""
		for j := 0; j < 12; j += 2 {
			x := string(res1[j+i]) + string(res1[i+j+1])
			HM_SK += ZWC[x]
		}
		s1 := s + HM_SK
		word[i/12] = s1 // Update the word array with the new value
		i += 12
	}

	outputFile, err := os.Create("stego_file.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	for _, w := range word {
		outputFile.WriteString(w + " ")
	}
	fmt.Println("\nStego file has successfully generated")
}

func encodeTxtData() {
	fileContent, err := ioutil.ReadFile("text_test.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	count2 := len(strings.Fields(string(fileContent)))
	bt := count2 / 6

	fmt.Println("Maximum number of words that can be inserted :- ", bt)

	var text1 string
	fmt.Print("\nEnter data to be encoded: ")
	fmt.Scanln(&text1)

	l := len(text1)
	if l <= bt {
		fmt.Println("\nInputted message can be hidden in the cover file\n")
		txtEncode(text1)
	} else {
		fmt.Println("\nString is too big please reduce string size")
	}
}

func decodeTxtData() {
	var stego string
	fmt.Print("\nPlease enter the stego file name(with extension) to decode the message: ")
	fmt.Scanln(&stego)

	fileContent, err := ioutil.ReadFile(stego)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	temp := ""
	for _, word := range strings.Fields(string(fileContent)) {
		binaryExtract := ""
		for _, letter := range word {
			if val, ok := ZWC[string(letter)]; ok {
				binaryExtract += val
			}
		}

        if binaryExtract == "111111111111" {
            break
        } else {
            temp += binaryExtract
        }
    }

    fmt.Println("\nEncrypted message presented in code bits:", temp)
    lengthd := len(temp)
    fmt.Println("\nLength of encoded bits:- ", lengthd)

    finalMsg := ""
    for i := 0; i < len(temp); i += 12 {
        t3, t4 := temp[i:i+4], temp[i+4:i+16]
        if t3 == "0110" {
            decimalData := binaryToDecimal(t4)
            finalMsg += string((decimalData ^ 170) + 48)
        } else if t3 == "0011" {
            decimalData := binaryToDecimal(t4)
            finalMsg += string((decimalData ^ 170) - 48)
        }
    }
    fmt.Println("\nMessage after decoding from the stego file:- ", finalMsg)
}

func binaryToDecimal(binary string) int {
	val, _ := strconv.ParseInt(binary, 2, 64)
	return int(val)
}

// Image Steganography Functions

func encodeImgData(img image.Image, data string) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
	    for x := bounds.Min.X; x < bounds.Max.X; x++ {
	        newImg.Set(x, y, img.At(x, y))
	    }
    }

	data += "*^*^*" // End marker for the hidden message
	binaryData := msgToBinary(data)

	indexData := 0

	for y := bounds.Min.Y; y < bounds.Max.Y && indexData < len(binaryData); y++ {
	    for x := bounds.Min.X; x < bounds.Max.X && indexData < len(binaryData); x++ {
	        r, g, b, a := newImg.At(x, y).RGBA()
	        r = (r &^ 1) | uint32(binaryData[indexData]-'0')
	        indexData++
	        if indexData < len(binaryData) {
	            g = (g &^ 1) | uint32(binaryData[indexData]-'0')
	            indexData++
	        }
	        if indexData < len(binaryData) {
	            b = (b &^ 1) | uint32(binaryData[indexData]-'0')
	            indexData++
	        }
	        newImg.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
	    }
	    if indexData >= len(binaryData) {
	        break
	    }
    }

	return newImg
}

func msgToBinary(msg string) []byte {
	var result []byte
	for _, char := range msg {
	    binaryChar := fmt.Sprintf("%08b", char)
	    result = append(result, []byte(binaryChar)...)
    }
	return result
}

func decodeImgData(img image.Image) string {
	bounds := img.Bounds()
	dataBinary := ""

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
	    for x := bounds.Min.X; x < bounds.Max.X; x++ {
	        r, g, b, _ := img.At(x, y).RGBA()
	        dataBinary += fmt.Sprintf("%d%d%d", r&1, g&1, b&1)
	    }
    }

	totalBytes := make([]string, (len(dataBinary)/8)+1)

	for i:=0; i<len(dataBinary); i+=8{
	    if i+8 <= len(dataBinary){
	        totalBytes = append(totalBytes, dataBinary[i:i+8])
	    }else{
	        totalBytes = append(totalBytes, dataBinary[i:])
	    }
    }

	var decodedMessage string

	for _, byteStr:= range totalBytes{
	    if byteStr == ""{
	        continue
	    }
	    charCode,_:= strconv.ParseInt(byteStr,2,64)
	    decodedMessage += string(charCode)

	    if strings.HasSuffix(decodedMessage,"*^*^*"){
	        return decodedMessage[:len(decodedMessage)-5]
	    }
    }

	return ""
}

// Audio Steganography Functions

func encodeAudio() {
	var audioPath string
	fmt.Print("Enter the path of the audio file to encode data into: ")
	fmt.Scanln(&audioPath)

	audioFile, err := os.Open(audioPath)
	if err != nil {
	    fmt.Println("Error opening audio file:", err)
	    return
    }
	defer audioFile.Close()

	reader := wav.NewDecoder(audioFile)

	var data string
	fmt.Print("Enter data to be encoded in the audio: ")
	fmt.Scanln(&data)

	data += "*^*^*" // End marker for the hidden message
	binaryData := msgToBinary(data)

	bufferSizeInSamplesPerChannel:=1024 // Number of samples to read at once.
	samplesPerChannel:=make([]int16,buffersizeInSamplesPerChannel)

	indexData:=0

	for{
    	n ,err:=reader.Read(samplesPerChannel[:]) // Read samples from audio.
    	if n==0 || err!=nil{
        	break // Exit loop when there are no more samples or an error occurs.
    	}

    	for i:=0;i<n;i++{
        	if indexData<len(binaryData){
            	samplesPerChannel[i]=samplesPerChannel[i]&^1|int16(binaryData[indexData]-'0') // Set LSB.
            	indexData++
        	  }else{
            	break // Exit loop when all data is encoded.
          }
         }

         // Write modified samples back to a new audio file.
         outAudioFile ,err:= os.Create("stego_audio.wav")
         if err!=nil{
             fmt.Println("Error creating output audio file:",err)
             return
         }

         writer:=wav.NewEncoder(outAudioFile,wav.SampleRate(reader.SampleRate()),reader.NumChannels(),16,wav.FormatPCM)
         defer outAudioFile.Close()

         writer.Write(samplesPerChannel[:n])
     }

     fmt.Println("Successfully encoded data into the audio.")
}

func decodeAudio() {
	var audioPath string
	fmt.Print("Enter the path of the stego audio file to decode data from: ")
	fmt.Scanln(&audioPath)

	audioFile ,err:= os.Open(audioPath)
	if err!=nil{
	    fmt.Println("Error opening audio file:",err)
	    return
    }
	defer audioFile.Close()

	reader:=wav.NewDecoder(audioFile)

	dataBinary:=""

	samplesPerChannel:=make([]int16,1024)

	for{
    	n ,err:=reader.Read(samplesPerChannel[:])
    	if n==0 || err!=nil{
    	    break
    	}

    	for i:=0;i<n;i++{
    	    dataBinary+=fmt.Sprintf("%d",samplesPerChannel[i]&1) // Extract LSB.
    	}
    }

	totalBytes:=make([]string,(len(dataBinary)/8)+1)

	for i:=0;i<len(dataBinary);i+=8{
    	if i+8<=len(dataBinary){
    	    totalBytes=append(totalBytes,dataBinary[i:i+8])
    	}else{
    	    totalBytes=append(totalBytes,dataBinary[i:])
    	}
    }

	var decodedMessage string

	for _,byteStr:=range totalBytes{
    	if byteStr==""{
    	    continue
    	}
    	charCode,_:=strconv.ParseInt(byteStr,2,64)
    	decodedMessage+=string(charCode)

    	if strings.HasSuffix(decodedMessage,"*^*^*"){
    	    return decodedMessage[:len(decodedMessage)-5] // Return message without end marker.
    	}
    }

	return ""
}

// Video Steganography Functions

func encodeVideo() {
	var videoPath string
	fmt.Print("Enter the path of the video file to encode data into: ")
	fmt.Scanln(&videoPath)

	videoCapture, err := gocv.VideoCapture(videoPath)
	if err != nil {
	    fmt.Println("Error opening video file:", err)
	    return
    }
	defer videoCapture.Close()

	var data string
	fmt.Print("Enter data to be encoded in the video: ")
	fmt.Scanln(&data)

	data += "*^*^*" // End marker for the hidden message
	binaryData := msgToBinary(data)

	frameWidth   : = int(videoCapture.Get(gocv.VideoCaptureWidth))
	frameHeight : = int(videoCapture.Get(gocv.VideoCaptureHeight))
	fourcc : = gocv.VideoWriterFourcc('M', 'J', 'P', 'G')
	outputVideo : = gocv.VideoWriter{}

	outputVideo.Open("stego_video.avi", fourcc, 30.0,
	image.Pt(frameWidth,
	frameHeight), true)

	frame : = gocv.NewMat()
	defer frame.Close()

	indexData : = 0

	for {
	    if ok : = videoCapture.Read(&frame); !ok {
	        break
        }

        for y : = 0; y < frameHeight && indexData < len(binaryData); y++ {
            for x : = 0; x < frameWidth && indexData < len(binaryData); x++ {
                pixelColor : = frame.GetVecbAt(y,
	x)
                pixelColor[0] & = ^uint8(1) | uint8(binaryData[indexData]-'0') // Modify LSB
                indexData++

                frame.SetVecbAt(y,
	x,
	pixelColor)
            }
        }

        outputVideo.Write(frame)
    }

    fmt.Println("Successfully encoded data into the video.")
}

func decodeVideo() {
	var videoPath string
	fmt.Print("Enter the path of the stego video file to decode data from: ")
	fmt.Scanln(&videoPath)

	videoCapture,
	err : = gocv.VideoCapture(videoPath)
	if err != nil {
	    fmt.Println("Error opening video file:", err)
	    return
    }
	defer videoCapture.Close()

	dataBinary : = ""
	frame : = gocv.NewMat()
	defer frame.Close()

	for {
	    if ok : = videoCapture.Read(&frame); !ok {
	        break
        }

        frameWidth : = frame.Rows()
        frameHeight : = frame.Cols()

        for y : = 0; y < frameHeight; y++ {
            for x : = 0; x < frameWidth; x++ {
                pixelColor : = frame.GetVecbAt(y,
	x)
                dataBinary + = fmt.Sprintf("%d", pixelColor[0]&uint8(1)) // Extract LSB
            }
        }
    }

	totalBytes : = make([]string,
(len(dataBinary)/8)+1)

	for i : = 0; i < len(dataBinary); i + = 8 {
	    if i + 8 <= len(dataBinary){
	        totalBytes=append(totalBytes,dataBinary[i:i+8])
        } else {
	        totalBytes=append(totalBytes,dataBinary[i:])
        }
    }

	var decodedMessage string

	for _, byteStr : = range totalBytes {
	    if byteStr == "" {
	        continue
        }
        charCode,
	err : = strconv.ParseInt(byteStr,
2,
64)
        decodedMessage + = string(charCode)

        if strings.HasSuffix(decodedMessage,"*^*^*") {
            return decodedMessage[:len(decodedMessage)-5] // Return message without end marker.
        }
    }

	return ""
}

// Main Function

func encodeImage() {
	var imgPath string
	fmt.Print("Enter the path of the image to encode data into: ")
	fmt.Scanln(&imgPath)

	imgFile, err := os.Open(imgPath)
	if err != nil {
	    fmt.Println("Error opening image file:", err)
	    return
    }
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
	    fmt.Println("Error decoding image:", err)
	    return
    }

	var data string
	fmt.Print("Enter data to be encoded in the image: ")
	fmt.Scanln(&data)

	newImg := encodeImgData(img,data)

	outputFileName :="stego_image.png"
	outFile ,err:= os.Create(outputFileName)
	if err!=nil{
	    fmt.Println("Error creating output file:",err)
	    return
    }

	defer outFile.Close()

	err = jpeg.Encode(outFile,newImg,nil)
	if err!=nil{
	    fmt.Println("Error encoding image:",err)
    }else{
        fmt.Println("Successfully encoded data into the image and saved as",outputFileName )
    }
}

func decodeImage() {
	var imgPath string
	fmt.Print("Enter the path of the stego image to decode data from: ")
	fmt.Scanln(&imgPath)

	imgFile ,err:= os.Open(imgPath)
	if err!=nil{
	    fmt.Println("Error opening image file:",err)
	    return
    }
	defer imgFile.Close()

	img ,_,err:= image.Decode(imgFile)
	if err!=nil{
	    fmt.Println("Error decoding image:",err)
	    return
    }

	message:=decodeImgData(img)
	if message==""{
	    fmt.Println("No hidden message found.")
    }else{
        fmt.Println("Decoded message from the image:",message)
    }
}

func main() {
	for {
    	fmt.Println("\n\t\t      STEGANOGRAPHY")
    	fmt.Println("1. Encode Text Message")
    	fmt.Println("2. Decode Text Message")
    	fmt.Println("3. Encode Image")
    	fmt.Println("4. Decode Image")
    	fmt.Println("5. Encode Audio")
    	fmt.Println("6. Decode Audio")
    	fmt.Println("7. Encode Video")
    	fmt.Println("8. Decode Video")
    	fmt.Println("9. Exit")
    	var choice int
    	fmt.Print("Enter your choice: ")
    	fmt.Scanln(&choice)

    	switch choice {
    	case 1:
        	encodeTxtData()
    	case 2:
        	decodeTxtData()
        case 3:
        	encodeImage()
        case 4:
        	decodeImage()
        case 5:
        	encodeAudio()
        case 6:
        	decodeAudio()
        case 7:
         encodeVideo()
       case 8:
          decodeVideo()
       case 9:
          return
       default:
          fmt.Println("Incorrect Choice")
      }
   }
}

