import { generateFingerprint } from "./fingerprint"
import { completeMultipartUpload, initMultipartUpload, uploadParts } from "./multipart"
import {partitionFile} from "./filePartition"

const startUploadVideo=async(selectedFile, setStatusMsg)=>{
    const fingerprint = await generateFingerprint(selectedFile)
    console.log("Fingerprint for the file: ",selectedFile.name+" is "+ fingerprint)

    // initialize multipart upload for the video
    
    const {msg, session} = await initMultipartUpload({
        name: selectedFile.name,
        type: selectedFile.type,
        size: selectedFile.size,
        fingerprint
    })

    console.log("INIT MULTIPART UPLOAD", msg+ "session ", session)

    if (session.status === "completed") {
        console.log("Video already uploaded")
        return { ...session, newSession: false }
    }

    let allParts = partitionFile(selectedFile, session.partSize)

    // already uploaded parts
    const uploadedPartNumbers = new Set(
        (session.uploadedParts || []).map(
            p => p.PartNumber
        )
    )

    // remaining parts
    const remainingParts = allParts.filter(
        part => !uploadedPartNumbers.has(
            part.partNumber
        )
    )

    let newlyUploadedParts = await uploadParts(session, remainingParts, () => {
        setStatusMsg(`Uploading parts...`)
    })

    for(const it of newlyUploadedParts){
      console.log("UPLOADED: ", it)
    }

    const allUploadedParts = [
        ...session.uploadedParts,
        ...newlyUploadedParts
    ].sort(
        (a, b) => a.PartNumber - b.PartNumber
    )

    const {msg:multipartUploadMsg, uploadID, videoID} = await completeMultipartUpload({
      uploadID: session.uploadID,
      key: session.key,
      videoID: session.videoID,
      parts: allUploadedParts
    })
    
    console.log("MULTIPART UPLOAD COMPLETE: ", multipartUploadMsg, videoID, uploadID)
    return {...session, newSession:true}
}

export default startUploadVideo