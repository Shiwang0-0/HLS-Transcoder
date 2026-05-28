import { useRef, useState, useEffect } from 'react'
import Button from './components/UploadBtn'
import { generatePresignedURL, notifyUploadComplete, uploadToS3 } from './helpers/s3'
import { waitForJobCompletion } from './helpers/pollHLS'
import VideoPlayer from './components/VideoPlayer'
import Spinner from './components/Spinner'
import Hls from 'hls.js'

const allowedTypes = {
  'video/mp4': true,
}

const App = () => {
  const fileInputRef = useRef(null)
  const videoRef = useRef(null)
  const hlsRef = useRef(null)

  const [selectedFile, setSelectedFile] = useState(null)
  const [videoMetadata, setVideoMetadata] = useState(null)
  const [streamURL, setStreamURL] = useState(null)
  const [status, setStatus] = useState(null)   // null | 'uploading' | 'transcoding' | 'error'
  const [statusMsg, setStatusMsg] = useState('')

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
    setStreamURL(null)
    setStatus(null)

    const video = document.createElement('video')
    video.preload = 'metadata'

    // after loading metadata, save the state
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
      URL.revokeObjectURL(video.src)
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
      setStatus('uploading')
      setStatusMsg('Uploading video to S3...')

      const { url, key, videoID, jobID } = await generatePresignedURL(videoMetadata)
      await uploadToS3(url, selectedFile, videoMetadata.type)
      await notifyUploadComplete(key, videoID, jobID)

      
      setStatus('transcoding')
      setStatusMsg('Transcoding in progress — waiting for stream to be ready...')
      
      // Poll until the .m3u8 manifest actually exists
      await waitForJobCompletion(jobID, (status, stage) => {
        setStatus(status)
        setStatusMsg(`Stage: ${stage}`)
      })
      const hlsURL = `${import.meta.env.VITE_HLS_BASE_URL}/${videoID}/master.m3u8`
      console.log('VideoID:', videoID)
      console.log('jobID:', jobID)
      console.log('Stream URL:', hlsURL)

      setStreamURL(hlsURL)
      setStatus(null)
      setStatusMsg('')
    } catch (err) {
      console.error('Failed:', err)
      setStatus('error')
      setStatusMsg(err.message)
    }
  }

  // on every change in streaming URL, change the hls instance
  useEffect(() => {
    if (!streamURL || !videoRef.current) return

    const video = videoRef.current

    const hls = new Hls()
    hlsRef.current = hls

    hls.loadSource(streamURL)
    hls.attachMedia(video)

    // start playback automatically
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      video.play()
    })

    return () => {
      hls.destroy()
      hlsRef.current = null
    }
  }, [streamURL])

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: '20px',
        padding: '40px',
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
          choice="choose"
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
        <Button
          btnName="Send"
          onClick={handleSend}
          choice="send"
          disabled={status === 'uploading' || status === 'transcoding'}
        />

        {/* Status indicator */}
        {status && status !== 'error' && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <Spinner />
            <p style={{ color: '#555', fontSize: '14px', margin: 0 }}>
              {statusMsg}
            </p>
          </div>
        )}

        {status === 'error' && (
          <p style={{ color: '#dc2626', fontSize: '14px', margin: 0 }}>
            {statusMsg}
          </p>
        )}

        {streamURL && <VideoPlayer src={streamURL} />}
      </div>
    </div>
  )
}

export default App