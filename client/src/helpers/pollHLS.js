export const streamJobStatus = (jobID, onProgress) => {
  const es = new EventSource(`http://localhost:8000/api/job/${jobID}/stream`)

  return new Promise((resolve, reject) => {
    let retryCount = 0
    const maxRetries = 5

    es.onmessage = (e) => {
      retryCount = 0 // real message arrived, connection is healthy

      const { status, stage } = JSON.parse(e.data)
      onProgress(status, stage)

      if (status === 'completed') { es.close(); resolve() }
      if (status === 'failed') { es.close(); reject(new Error(stage || 'Transcoding failed')) }
    }

    es.onerror = () => {
      // CONNECTING = browser is auto-retrying
      if (es.readyState === EventSource.CONNECTING) {
        retryCount++
        if (retryCount > maxRetries) {
          es.close()
          reject(new Error('Connection lost after multiple retries'))
        }
        return
      }

      // CLOSED = browser has permanently given up
      es.close()
      reject(new Error('Connection lost'))
    }
  })
}