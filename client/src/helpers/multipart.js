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
            throw new Error(errorData.error || 'Failed to initialize upload')
        }

        const data = await response.json()

        return data
}

export const uploadParts=async(initData, allParts)=>{
    const uploadedParts = await Promise.all(
        allParts.map(async (it) => {

            const { url, msg } = await generatePresignedPartURL({
                uploadID: initData.uploadID,
                objectKey: initData.key,
                partNumber: it.partNumber
            })
            console.log("PRESIGNED URL: "+it.partNumber+" "+ msg)

            return uploadPartToS3(url, it.part, it.partNumber)
        })
    )
    return uploadedParts
}

export const completeMultipartUpload =async ({ uploadID, key, videoID, parts}) => {

    const response = await fetch(
      'http://localhost:8000/api/complete-multipart-upload',
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },

        body: JSON.stringify({ uploadID, key, videoID, parts})
      }
    )

    const data = await response.json()

    if (!response.ok) {
        throw new Error(data.error || data.msg || 'Failed to complete upload')
    }

    return data
}