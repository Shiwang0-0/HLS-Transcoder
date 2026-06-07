export const generatePresignedPartURL = async (payload)=>{
    try{
         const response = await fetch('http://localhost:8000/api/presigned-part-url',{
            method:'POST',
            headers:{
                'Content-Type':'application/json',
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
        throw err;
    }
}

export const uploadPartToS3 = async (presignedPartURL,part,PartNumber) => {

    // resume upload test
    // if (PartNumber == 2)
        // throw new Error(`Simulated failure for part ${PartNumber}`)
    
    // console.log("presigned URL:", presignedPartURL)
    const response = await fetch(presignedPartURL,
        {
            method: "PUT",
            body: part,
        }
    )

    if (!response.ok) {
        throw new Error(
        `Failed to upload part ${PartNumber}`
        )
    }
    const ETag = response.headers.get("ETag")

    if (!ETag) {
        throw new Error(
        `Missing ETag for part ${PartNumber}`
        )
    }

    console.log("UPLOADED PARTNUMBER: "+PartNumber+" \n")

    return {PartNumber,ETag}
}

export const verifyAndPersistParts = async (sessionID, parts) => {
    const response = await fetch(`http://localhost:8000/api/uploads/${sessionID}/parts`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(parts)
    })

    if (!response.ok) {
        throw new Error('Failed to verify parts')
    }

    const data = await response.json()

    return data
}

export const createTranscodingJob = async(key, videoID )=>{
    try{
        const payload={
            key,
            videoID,
        }

        const response = await fetch('http://localhost:8000/api/job',{
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