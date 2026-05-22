import React from 'react'
import Button from './components/UploadBtn'
import VideoPlayer from './components/VideoPlayer'

const App = () => {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: '20px',
        padding: '40px',
        maxWidth: '800px',
        margin: 'auto',
      }}
    >
      <VideoPlayer />
      <Button />
    </div>
  )
}

export default App