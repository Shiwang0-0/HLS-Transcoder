import { generatePresignedPartURL, uploadPartToS3 } from "./s3"

export const initMultipartUpload= async({name, type, size, fingerprint})=>{
        const payload = {
            name,
            type,
            size,
            fingerprint
        }
        const response = await fetch('http://localhost:8000/api/init-multipart-upload',
            {
                method: 'POST',
                headers: {
                'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            }
        )

        if (!response.ok) {
            const errorData = await response.json()
            throw new Error(errorData.msg || 'Failed to initialize upload')
        }

        const data = await response.json()

        return data
}

export const uploadParts=async(initData, allParts)=>{
    const result = await Promise.allSettled(
        allParts.map(async (it) => {

            const { url } = await generatePresignedPartURL({
                sessionID: initData.id,
                PartNumber: it.PartNumber
            })

            return uploadPartToS3(url, it.part, it.PartNumber)
        })
    )

    // for all the promises that were fullfilled, save them
    const uploadedParts = result.filter(r=>r.status === 'fulfilled' && r.value!=null).map(r=>r.value)
    const failedPartNumbers = allParts
        .filter((_, i) => result[i].status === 'rejected' || result[i].value == null)
        .map(p => p.PartNumber)

    if (failedPartNumbers.length > 0) {
        console.warn(`Parts failed to upload: ${failedPartNumbers}`)
    }

    return { uploadedParts, failedPartNumbers }
}

export const completeMultipartUpload =async ({ sessionID, videoID, parts}) => {

    const response = await fetch(
      'http://localhost:8000/api/complete-multipart-upload',
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },

        body: JSON.stringify({ sessionID, videoID, parts})
      }
    )

    const data = await response.json()

    if (!response.ok) {
        throw new Error(data.error || data.msg || 'Failed to complete upload')
    }

    return data
}