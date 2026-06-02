import { useState, useEffect } from "react"
import VideoPlayer from "./components/VideoPlayer"
import { UploadModal } from "./components/Upload"
import { VideoCard } from "./components/VideoCard"


const ACTIVE_STATUSES = ["transcoding", "downloading", "uploading", "queued"]

export default function VideoFeed() {
  const [videos, setVideos] = useState([])
  const [activeVideo, setActiveVideo] = useState(null)
  const [filter, setFilter] = useState("all")
  const [loading, setLoading] = useState(true) 
  const [showUpload, setShowUpload] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    let isMounted = true

    const fetchVideos = async () => {
      try {
        const res = await fetch("http://localhost:8000/api/jobs")
        if (!res.ok) throw new Error()

        const data = await res.json()
         if (isMounted) {
          const jobsList = data.jobs || data
          setVideos(jobsList)

          // Keep the currently selected active video in sync if its status changes in the background
          if (activeVideo) {
            const freshActiveVideo = jobsList.find(v => v.jobID === activeVideo.jobID)
            if (freshActiveVideo && freshActiveVideo.status !== activeVideo.status) {
              setActiveVideo(freshActiveVideo)
            }
          }
        }
      } catch {
        if (isMounted) setVideos([])
      } finally {
        if (isMounted) setLoading(false)
      }
    }

    fetchVideos()
    return () => { isMounted = false }
  }, [refreshKey,activeVideo])

  useEffect(() => {
    const safeVideos = Array.isArray(videos) ? videos : []
    const hasActiveJobs = safeVideos.some(v => ACTIVE_STATUSES.includes(v.status))

    if (!hasActiveJobs) return // If no jobs are active, don't set up an interval timer

    const intervalId = setInterval(() => {
      setRefreshKey(prev => prev + 1)
    }, 3000)

    return () => clearInterval(intervalId) // Clean up the timer when statuses change or component unmounts
  }, [videos])

  const handleUploadComplete = () => {
    setShowUpload(false)
    setLoading(true) 
    setRefreshKey(prev => prev + 1) 
  }

  const filters = [
    { key: "all",        label: "all" },
    { key: "completed",  label: "ready" },
    { key: "processing", label: "processing" },
    { key: "failed",     label: "failed" },
  ]

  const allVideos = Array.isArray(videos) ? videos : []

  const filtered = allVideos.filter(v => {
    return filter === "all" ||
      (filter === "completed"  && v.status === "completed") ||
      (filter === "processing" && ["transcoding", "downloading", "uploading", "queued"].includes(v.status)) ||
      (filter === "failed"     && v.status === "failed")
  })

  const activeVideoSrc = activeVideo
    ? `${import.meta.env.VITE_HLS_BASE_URL}/${activeVideo.videoID}/master.m3u8`
    : null

  return (
    <div className="min-h-screen bg-[#080808] font-sans text-[#e5e5e5]">
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500&family=DM+Mono:wght@400;500&display=swap');
        ::-webkit-scrollbar { width: 4px; }
        ::-webkit-scrollbar-track { background: #0d0d0d; }
        ::-webkit-scrollbar-thumb { background: #222; border-radius: 2px; }
      `}</style>

      {showUpload && <UploadModal onClose={handleUploadComplete} />}

      <div className="border-b border-[#151515] px-6 py-3.5 flex items-center justify-between sticky top-0 bg-[#080808] z-10">
        <div className="flex items-center gap-2.5">
          <div className="w-6 h-6 rounded bg-[#3b82f6] flex items-center justify-center">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="white">
              <polygon points="5,3 19,12 5,21" />
            </svg>
          </div>
          <span className="text-[15px] font-medium tracking-tight">streamvault</span>
        </div>
        <button
          onClick={() => setShowUpload(true)}
          className="bg-[#3b82f6] border-none rounded-md color-white cursor-pointer px-3.5 py-1.5 text-[12px] font-sans font-medium"
        >
          + upload
        </button>
      </div>

      <div className="max-w-[1100px] mx-auto px-6 py-6">
        {activeVideo && (
          <div className="bg-[#0d0d0d] border border-[#222] rounded-xl overflow-hidden mb-6 animate-slide-down">
            <div className="flex items-center justify-between px-4 py-2.5 border-b border-[#1a1a1a]">
              <div className="flex items-center gap-2.5">
                <div className="w-1.5 h-1.5 rounded-full bg-[#3b82f6]" />
                <span className="text-[13px] font-medium text-[#e5e5e5]">
                  {(activeVideo.videoName || "").replace(/\.(mp4|mov|avi|mkv)$/i, "")}
                </span>
              </div>
              <button
                onClick={() => setActiveVideo(null)}
                className="bg-transparent border border-[#222] rounded text-[#666] cursor-pointer px-2 py-0.5 text-[11px] font-mono"
              >close</button>
            </div>
            <div className="p-4">
              <VideoPlayer src={activeVideoSrc} />
            </div>
          </div>
        )}

        <div className="flex items-center gap-3 mb-6 flex-wrap">
          <div className="flex gap-1.5">
            {filters.map(f => (
              <button 
                key={f.key} 
                onClick={() => setFilter(f.key)} 
                className={`
                  rounded-md px-3 py-1.5 text-[12px] cursor-pointer font-sans transition-all
                  ${filter === f.key ? "bg-[#1e1e1e] border border-[#333] text-[#e5e5e5]" : "bg-transparent border border-[#1a1a1a] text-[#555]"}
                `}
              >
                {f.label}
              </button>
            ))}
          </div>
          <div className="ml-auto text-[12px] text-[#333] font-mono">
            {filtered.length} results
          </div>
        </div>

        {loading ? (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="bg-[#111] border border-[#1a1a1a] rounded-lg overflow-hidden animate-pulse">
                <div className="w-full aspect-video bg-[#0d0d0d]" />
                <div className="p-3">
                  <div className="h-3 bg-[#1a1a1a] rounded w-4/5 mb-2" />
                  <div className="h-2.5 bg-[#161616] rounded w-1/2" />
                </div>
              </div>
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-16 px-4 text-[#333] animate-fade-in">  
            <p className="text-[13px] font-mono">no videos found</p>
          </div>
        ) : (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-4 animate-fade-in">
            {filtered.map(v => (
              <VideoCard
                key={v.jobID} video={v}
                isActive={activeVideo?.jobID === v.jobID}
                onClick={setActiveVideo}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}