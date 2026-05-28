import { useEffect, useRef, useState } from 'react'
import Hls from 'hls.js'

const VideoPlayer = ({ src }) => {
  const videoRef = useRef(null)
  const hlsRef = useRef(null)
  const [levels, setLevels] = useState([])       // available quality levels
  const [currentLevel, setCurrentLevel] = useState(-1)  // -1 = auto

  useEffect(() => {
    const video = videoRef.current
    if (!video || !src) return

    if (Hls.isSupported()) {
      const hls = new Hls({
        maxLoadingRetry: 6,
        manifestLoadingRetryDelay: 2000,
        manifestLoadingMaxRetryTimeout: 64000,
        levelLoadingRetryDelay: 2000,
        fragLoadingRetryDelay: 2000,
      })
      hlsRef.current = hls

      // load master.m3u8
      hls.loadSource(src)
      hls.attachMedia(video)

      hls.on(Hls.Events.MANIFEST_PARSED, (event, data) => {
        // data.levels is an array of { height, bitrate, ... }
        setLevels(data.levels)
        setCurrentLevel(-1) // start on auto
        video.play().catch((err) => console.warn('Autoplay blocked:', err))
      })

      hls.on(Hls.Events.LEVEL_SWITCHED, () => {
        // Sync UI if ABR auto-switches
        if (hls.autoLevelEnabled) setCurrentLevel(-1)
      })

      hls.on(Hls.Events.ERROR, (_, data) => {
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              hls.startLoad()
              break
            case Hls.ErrorTypes.MEDIA_ERROR:
              hls.recoverMediaError()
              break
            default:
              hls.destroy()
          }
        }
      })

      return () => hls.destroy()
    }

    // Safari native HLS
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = src
      video.addEventListener('loadedmetadata', () => {
        video.play().catch((err) => console.warn('Autoplay blocked:', err))
      })
    }
  }, [src])

  const handleQualityChange = (levelIndex) => {
    const hls = hlsRef.current
    if (!hls) return

    // -1 tells hls.js to resume automatic ABR
    hls.currentLevel = levelIndex
    setCurrentLevel(levelIndex)
  }

  const qualityLabel = (level) => `${level.height}p`

  return (
    <div style={{ width: '100%', maxWidth: '1600px', margin: '0 auto' }}>
      <video
        ref={videoRef}
        controls
        width="100%"
        style={{
          borderRadius: '12px',
          marginTop: '20px',
          backgroundColor: '#000',
          width: '100%',
          height: '600px',
          objectFit: 'contain',
        }}
      />
      {levels.length > 0 && (
        <div
          style={{
            marginTop: '18px',
            display: 'flex',
            gap: '8px',
            justifyContent: 'center',
            alignItems: 'center',
            width: '100%',
          }}
        >
          <QualityBtn
            label="Auto"
            active={currentLevel === -1}
            onClick={() => handleQualityChange(-1)}
          />
          {levels.map((level, index) => (
            <QualityBtn
              key={index}
              label={qualityLabel(level)}
              active={currentLevel === index}
              onClick={() => handleQualityChange(index)}
            />
          ))}
        </div>
      )}
    </div>
  )
}

const QualityBtn = ({ label, active, onClick }) => (
  <button
    onClick={onClick}
    style={{
      padding: '4px 12px',
      borderRadius: '6px',
      border: '1px solid #4f46e5',
      background: active ? '#4f46e5' : 'transparent',
      color: active ? '#fff' : '#4f46e5',
      cursor: 'pointer',
      fontSize: '13px',
      fontWeight: active ? 600 : 400,
    }}
  >
    {label}
  </button>
)

export default VideoPlayer