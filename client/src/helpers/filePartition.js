export const getRemainingParts = (file, partSize, uploadedPartNumbersSet) => {
    const remainingParts = [];
    let start = 0;
    let PartNumber = 1;

    while (start < file.size) {
        // ONLY slice and store if we actually need to upload this part
        if (!uploadedPartNumbersSet.has(PartNumber)) {
            const part = file.slice(start, start + partSize);
            remainingParts.push({ PartNumber, part });
        }

        start += partSize;
        PartNumber++;
    }

    return remainingParts;
};