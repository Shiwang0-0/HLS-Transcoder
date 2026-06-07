import SparkMD5 from 'spark-md5';

export const generateFingerprint = (file) => {
    return new Promise((resolve, reject) => {
        const chunkSize = 2 * 1024 * 1024; // Read in 2MB chunks
        const chunks = Math.ceil(file.size / chunkSize);
        let currentChunk = 0;
        
        const spark = new SparkMD5.ArrayBuffer();
        const fileReader = new FileReader();

        fileReader.onload = (e) => {
            // Append this specific chunk to the running hash calculation
            spark.append(e.target.result); 
            currentChunk++;

            if (currentChunk < chunks) {
                loadNext();
            } else {
                // return the final hex string
                resolve(spark.end());
            }
        };

        fileReader.onerror = () => {
            reject(new Error("File reading failed during fingerprint generation"));
        };

        const loadNext = () => {
            const start = currentChunk * chunkSize;
            const end = Math.min(start + chunkSize, file.size);
            fileReader.readAsArrayBuffer(file.slice(start, end));
        };

        loadNext();
    });
};