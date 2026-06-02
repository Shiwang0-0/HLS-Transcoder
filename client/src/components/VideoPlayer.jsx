import { useEffect, useRef, useState } from 'react'
import Hls from 'hls.js'

const VideoPlayer = ({ src }) => {
  const videoRef = useRef(null)
  const hlsRef = useRef(null)
  const [levels, setLevels] = useState([])       
  const [currentLevel, setCurrentLevel] = useState(-1)  

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

      hls.loadSource(src)
      hls.attachMedia(video)

      hls.on(Hls.Events.MANIFEST_PARSED, (event, data) => {
        setLevels(data.levels)
        setCurrentLevel(-1) 
        video.play().catch((err) => console.warn('Autoplay blocked:', err))
      })

      hls.on(Hls.Events.LEVEL_SWITCHED, () => {
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

    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = src
      video.addEventListener('loadedmetadata', () => {
        video.play().catch((err) => console.warn('Autoplay blocked:', err))
      })
    }
  }, [src])

  const handleQualityChange = (levelIndex) => {
    const hls = hlsRef.current
    const video = videoRef.current
    if (!hls || !video) return

    if (levelIndex === -1) {
      hls.loadLevel = -1 
      setCurrentLevel(-1)
    } else {
      setCurrentLevel(levelIndex)
      hls.nextLevel = levelIndex

      if (!video.paused) {
        video.pause()
        setTimeout(() => {
          video.play().catch((err) => console.warn('Playback resume failed:', err))
        }, 150) 
      }
    }
  }

  const qualityLabel = (level) => `${level.height}p`

  return (
    <div className="w-full max-w-[1600px] mx-auto">
      <video
        ref={videoRef}
        controls
        className="w-full h-[600px] rounded-xl mt-5 bg-black object-contain"
      />
      {levels.length > 0 && (
        <div className="mt-4.5 flex gap-2 justify-center items-center w-full">
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
    className={`
      px-3 py-1 rounded-md border text-sm cursor-pointer transition-colors duration-150
      ${active 
        ? 'border-[#4f46e5] bg-[#4f46e5] text-white font-semibold' 
        : 'border-[#4f46e5] bg-transparent text-[#4f46e5] font-normal'
      }
    `}
  >
    {label}
  </button>
)

export default VideoPlayer