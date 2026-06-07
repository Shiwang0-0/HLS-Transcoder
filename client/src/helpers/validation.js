export function validateFile(file) {
    const allowedTypes = ["video/mp4"]
    const maxSize = 500 * 1024 * 1024 // 500 MB

    if (!file?.name?.trim()) {
        return "File name is required"
    }

    if (!allowedTypes.includes(file.type)) {
        return "Media type not allowed, try .mp4"
    }

    if (file.size <= 0) {
        return "Invalid file size"
    }

    if (file.size > maxSize) {
        return "File size cannot exceed 500MB"
    }

    return null
}