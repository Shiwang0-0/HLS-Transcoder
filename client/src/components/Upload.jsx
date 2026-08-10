import { useRef, useState } from 'react'
import Button from './UploadBtn'
import Spinner from './Spinner'
import startUploadVideo from '../helpers/upload'
import { createTranscodingJob } from '../helpers/s3'
import { streamJobStatus } from '../helpers/pollHLS'
import {validateFile} from "../helpers/validation"


const Upload = ({onUploadComplete}) => {
  const fileInputRef = useRef(null)

  const [selectedFile, setSelectedFile] = useState(null)
  const [status, setStatus] = useState(null)   
  const [statusMsg, setStatusMsg] = useState('')

  const handleChooseClick = () => {
    fileInputRef.current.click()
  }

  const handleFileChange = (e) => {
    const file = e.target.files[0]
    if (!file) return

    const error = validateFile(file)

    if (error) {
        alert(error)
        e.target.value = ""
        return
    }

    setSelectedFile(file)
    setStatus(null)
  }

  const handleSend = async () => {
    if (!selectedFile) return

    try {
      setStatus('uploading')
      setStatusMsg('Uploading video to S3...')

      const session = await startUploadVideo(selectedFile, setStatusMsg)

      const streamPromise = streamJobStatus(session.jobID, (status, stage) => {
        setStatus(status)
        setStatusMsg(`Stage: ${stage}`)
      })

      if(session.newSession){
        setStatusMsg('Pushing job to queue...')
        const { msg: queueMsg, job } = await createTranscodingJob(session.id, session.videoID)
        console.log('QUEUE RESPONSE:', queueMsg, job)
      }

      await streamPromise

      setStatus('success')
      setStatusMsg('Your video was uploaded and processed successfully!')

    } catch (err) {
      console.error('Failed:', err)
      setStatus('error')
      setStatusMsg(err.message)
    }
  }

  
  return (
    <div className="flex flex-col gap-5 p-6 m-auto w-full">
      <input type="file" ref={fileInputRef} className="hidden" onChange={handleFileChange} />

      {status === 'success' ? (
        <div className="flex flex-col items-center justify-center text-center p-6 bg-[#0f172a]/30 border border-emerald-950/50 rounded-xl animate-fade-in">
          <div className="w-12 h-12 rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center mb-4 text-emerald-400">
            <svg height={24} width={24} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2.5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h3 className="text-base font-semibold text-white mb-1">Upload Complete!</h3>
          <p className="text-sm text-neutral-400 max-w-sm leading-relaxed mb-6">
            "{selectedFile?.name.replace(/\.(mp4)$/i, '')}" is now fully processed and ready to watch.
          </p>
          <button
            onClick={onUploadComplete}
            className="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 transition-colors text-white text-sm font-semibold rounded-lg shadow-lg shadow-emerald-900/20 cursor-pointer"
          >
            Done
          </button>
        </div>
      ) : (
        <>
          <div className="flex flex-col items-center justify-center">
            <Button btnName="Choose File" onClick={handleChooseClick} choice="choose" disabled={status && status !== 'error'} />
            {selectedFile && (
              <p className="text-neutral-400 text-sm mt-3 font-mono break-all text-center max-w-xs">
                Selected: {selectedFile.name}
              </p>
            )}
          </div>

          <div className="flex flex-col items-center justify-center gap-4 w-full">
            <Button
              btnName="Send"
              onClick={handleSend}
              choice="send"
              disabled={!selectedFile || status === 'uploading' || status === 'queuing' || status === 'transcoding'}
            />

            {status && status !== 'error' && status !== 'success' && (
              <div className="flex items-center gap-3 bg-neutral-900/50 border border-neutral-800 px-4 py-2.5 rounded-lg animate-pulse">
                <Spinner />
                <p className="text-neutral-300 text-sm font-medium tracking-wide m-0">
                  {statusMsg}
                </p>
              </div>
            )}

            {status === 'error' && (
              <div className="text-center bg-red-950/20 border border-red-900/30 px-4 py-3 rounded-lg w-full max-w-sm">
                <p className="text-red-400 text-sm font-medium m-0">
                  {statusMsg}
                </p>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}

export function UploadModal({ onClose }) {
  return (
    <div
      onClick={e => e.target === e.currentTarget && onClose()}
      className="fixed inset-0 z-[100] bg-black/75 backdrop-blur-[4px] flex items-center justify-center transition-all animate-fade-in"
    >
      <div className="bg-[#0d0d0d] border border-[#222] rounded-xl w-full max-w-[520px] p-6 relative transition-all">
        <div className="flex items-center justify-between mb-5">
          <span className="text-[14px] font-medium text-[#e5e5e5]">upload video</span>
          <button onClick={onClose} className="bg-transparent border border-[#222] rounded text-[#666] cursor-pointer px-2 py-0.5 text-[11px] font-mono">
            close
          </button>
        </div>
        <Upload onUploadComplete={onClose} />
      </div>
    </div>
  )
}

export default Upload