export const partitionFile=(file, partSize)=>{
    const parts = []

    let start = 0
    let partNumber = 1

    while (start < file.size) {
        const part = file.slice(start,start + partSize)

        parts.push({partNumber,part})

        start += partSize
        partNumber++
    }

    return parts
}