export const waitForJobCompletion = async (jobID, onProgress) => {
  const intervalMs = 3000

  while (true) {
    const res = await fetch(`http://localhost:8000/api/job/${jobID}`)
    const job = await res.json()

    if (onProgress) {
      onProgress(job.status, job.stage)
    }

    if (job.status === "completed") {
      return
    }

    if (job.status === "failed") {
      throw new Error(job.error || "Transcoding failed")
    }

    await new Promise(r => setTimeout(r, intervalMs))
  }
}