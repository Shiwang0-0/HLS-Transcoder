import { useRef, useState } from 'react'
import Button from './components/UploadBtn'
import { generatePresignedURL, notifyUploadComplete, uploadToS3 } from './helpers/s3'

const allowedTypes = {
  'video/mp4': true,
}

const App = () => {
  const fileInputRef = useRef(null)

  const [selectedFile, setSelectedFile] = useState(null)
  const [videoMetadata, setVideoMetadata] = useState(null)

  const handleChooseClick = () => {
    fileInputRef.current.click()
  }

  const handleFileChange = (e) => {
    const file = e.target.files[0]
    if (!file) return

    if (!allowedTypes[file.type]) {
      alert('Unsupported video format')
      return
    }

    setSelectedFile(file)

    const video = document.createElement('video')
    video.preload = 'metadata'

    video.onloadedmetadata = () => {
      setVideoMetadata({
        name: file.name,
        size: file.size,
        type: file.type,
        duration: video.duration,
        width: video.videoWidth,
        height: video.videoHeight,
        lastModified: file.lastModified,
      })
      URL.revokeObjectURL(video.src) // cleanup after metadata extraction
    }

    video.onerror = () => {
      alert('Failed to load video metadata')
    }

    video.src = URL.createObjectURL(file)
  }

  const handleSend = async () => {
    if (!selectedFile) {
      alert('Choose a file first')
      return
    }

    if (!videoMetadata) {
      alert('Error loading video metadata')
      return
    }

    try {
      const { url, key } = await generatePresignedURL(videoMetadata)
      await uploadToS3(url, selectedFile, videoMetadata.type)
      // make a call to backend to notify that upload is complete, so info can be pushed to sqs
      await notifyUploadComplete(key)
    } catch (err) {
      console.error('Upload failed:', err)
      alert('Upload failed: ' + err.message)
    }
  }

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: '20px',
        padding: '40px',
        maxWidth: '600px',
        margin: 'auto',
      }}
    >
      <input
        type="file"
        ref={fileInputRef}
        style={{ display: 'none' }}
        onChange={handleFileChange}
      />

      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          flexDirection: 'column',
        }}
      >
        <Button
          btnName="Choose File"
          onClick={handleChooseClick}
          variant="choose"
        />

        {selectedFile && (
          <p style={{ color: 'black', fontSize: '14px', marginTop: '10px' }}>
            Selected: {selectedFile.name}
          </p>
        )}
      </div>

      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          flexDirection: 'column',
          gap: '1rem',
        }}
      >
        <Button btnName="Send" onClick={handleSend} variant="send" />
      </div>
    </div>
  )
}

export default App