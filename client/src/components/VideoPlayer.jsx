import React, { useEffect, useRef } from 'react'
import videojs from 'video.js'
import 'video.js/dist/video-js.css'

const VideoPlayer = () => {
  const videoRef = useRef(null)
  const playerRef = useRef(null)

  useEffect(() => {
    if (!videoRef.current) return

    const videoElement = document.createElement('video-js')

    videoElement.classList.add('vjs-big-play-centered')

    videoRef.current.appendChild(videoElement)

    const player = (playerRef.current = videojs(videoElement, {
      controls: true,
      responsive: true,
      fluid: true,
      sources: [
        {
          src: 'bocchi.mp4',
          type: 'video/mp4',
        },
      ],
    }))

    return () => {
      if (player && !player.isDisposed()) {
        player.dispose()
      }
    }
  }, [])

  return (
    <div data-vjs-player>
      <div ref={videoRef} />
    </div>
  )
}

export default VideoPlayer