import { generateFingerprint } from "./fingerprint"
import { completeMultipartUpload, initMultipartUpload, uploadParts } from "./multipart"
import {getRemainingParts} from "./filePartition"
import { verifyAndPersistParts } from "./s3"

const startUploadVideo=async(selectedFile, setStatusMsg)=>{

    setStatusMsg(`Generating file hash...`)
    const fingerprint = await generateFingerprint(selectedFile)
    console.log("Fingerprint for the file: ",selectedFile.name+" is "+ fingerprint)

    // initialize multipart upload for the video
    
    const {msg, session} = await initMultipartUpload({
        name: selectedFile.name,
        type: selectedFile.type,
        size: selectedFile.size,
        fingerprint
    })

    console.log("INIT MULTIPART UPLOAD", msg)

    if (session.status === "completed") {
        console.log("Video already uploaded")
        return { ...session, newSession: false }
    }

    if (session.status === 'uploading' && session.uploadedParts.length > 0) {
        setStatusMsg(`Resuming previous upload...`)
    }

    // already uploaded parts
    let uploadedPartNumbers = new Set(
        (session.uploadedParts || []).map(p => p.PartNumber)
    );

    // Get only the parts we need to process in a single pass
    let remainingParts = getRemainingParts(selectedFile, session.partSize, uploadedPartNumbers);

    const MAX_RETRIES = 5;

    for (let attempt = 1; attempt <= MAX_RETRIES; attempt++) {

        if (remainingParts.length === 0) 
            break
         
        setStatusMsg(`Uploading parts...`)

        const {uploadedParts: newlyUploadedParts, failedPartNumbers} = await uploadParts(session, remainingParts)

        // console.log("verfying & persisting session", session.id, newlyUploadedParts)
        const {verifiedParts, missingParts} = await verifyAndPersistParts(session.id, newlyUploadedParts)
        // console.log(`attempt ${attempt}: verified: ${verifiedParts.length} & missing: ${missingParts.length}`)

        const allMissing = [
            ...missingParts,
            ...failedPartNumbers.map(pn => ({ PartNumber: pn }))
        ]

        if(allMissing.length === 0){ // all were uploaded successfully
            remainingParts = []
            break;
        }

        let missingPartNumbers=new Set(
            allMissing.map(p=>p.PartNumber)
        )

        
        uploadedPartNumbers = new Set([
            ...uploadedPartNumbers, // already uploaded
            ...(verifiedParts.map(p=>p.PartNumber)) // just uploaded and verfied
        ]
        )
        remainingParts = getRemainingParts(selectedFile, session.partSize, uploadedPartNumbers) // includes those are neither in missing nor in verified because they were removed in previous iterations
        remainingParts = remainingParts.filter(p=>missingPartNumbers.has(p.PartNumber)) // should belong to the missing parts

        // console.warn(`Re-uploading ${allMissing.length} missing parts:`, missingParts)
    
    
        // for(const it of newlyUploadedParts){
        //   console.log("UPLOADED: ", it)
        // }
    }

     if (remainingParts.length > 0) {
        throw new Error(`Upload failed after ${MAX_RETRIES} attempts`)
    }


    const {verifiedParts: allVerifiedParts} = await verifyAndPersistParts(session.id, []) // verify that are actually uploaded to the server
    
    const sortedParts =  [...allVerifiedParts].sort(
        (a, b) => a.PartNumber - b.PartNumber
    )

    const {msg:multipartUploadMsg} = await completeMultipartUpload({
      sessionID: session.id,
      videoID: session.videoID,
      parts: sortedParts
    })
    
    console.log("MULTIPART UPLOAD COMPLETION MSG: ", multipartUploadMsg)
    return {...session, newSession:true}
}

export default startUploadVideo