// we are loading the whole file, instead running hash for different chunks can be used but since we are not dealing with huge files, it is okay for now
export const generateFingerprint = async (file) => {
  const buffer = await file.arrayBuffer()

  // Generate SHA-256 hash
  const hashBuffer = await crypto.subtle.digest('SHA-256',buffer)

  // convert hash bytes -> hex string
  const hashArray = Array.from(new Uint8Array(hashBuffer))

  const hashHex = hashArray
    .map(byte => byte.toString(16).padStart(2, '0'))
    .join('')

  return hashHex
}