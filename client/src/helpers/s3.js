export const generatePresignedURL=async (videoMetadata)=>{
    try{
        const response = await fetch('http://localhost:8000/api/presigned-url',{
            method:'POST',
            headers:{
                'Content-Type':'application/json',
            },
            body:JSON.stringify(videoMetadata)
        })

        if(!response.ok){
            const errorData = await response.json()
            console.log(errorData)

            throw new Error(errorData.error || "Unknown error")
        }

        const data = await response.json()
        return data
    }catch(err){
        console.log(err)
    }
}

export const uploadToS3 = async(presignedURL, file, fileType)=>{
    const response = await fetch(presignedURL, {
        method: "PUT",
        body: file,
        headers: {
            "Content-Type": fileType,
        },
    })

    if (!response.ok) {
        throw new Error("Failed to upload to S3")
    }

    return response
}

export const notifyUploadComplete = async(key)=>{
    try{
        const payload={
            key
        }

        const response = await fetch('http://localhost:8000/api/notify-upload',{
            method:'POST',
            headers:{
                'Content-Type':'application/json'
            },
            body:JSON.stringify(payload)
        })

        if(!response.ok){
                const errorData = await response.json()
                console.log(errorData)

                throw new Error(errorData.error || "Unknown error")
            }

        const data = await response.json()
        return data
    }catch(err){
        console.log(err)
    }
}