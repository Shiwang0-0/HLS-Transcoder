export const waitForJobCompletion = async (jobID, onProgress) => {
  const intervalMs = 3000

  while (true) {
    const res = await fetch(`http://localhost:8000/api/job/${jobID}`)
    const data = await res.json()
    const job=data.job

    // console.log("job: ", job)
    console.log("status:", JSON.stringify(job.status)) 

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