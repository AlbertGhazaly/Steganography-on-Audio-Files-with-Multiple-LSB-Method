export class ValidationService {
    static validateKey(input) {
        let value = input.value;
        let hasInvalidChars = false;
        
        let filteredValue = '';
        for (let i = 0; i < value.length; i++) {
            const charCode = value.charCodeAt(i);
            if (charCode <= 255) {
                filteredValue += value.charAt(i);
            } else {
                hasInvalidChars = true;
            }
        }
        
        input.value = filteredValue;
        
        const parent = input.parentElement;
        const existingError = parent.querySelector('.key-error');
        
        if (hasInvalidChars) {
            if (!existingError) {
                const errorDiv = document.createElement('div');
                errorDiv.className = 'key-error text-red-500 text-sm mt-1';
                errorDiv.textContent = 'Key can only contain ASCII characters (0-255)';
                parent.appendChild(errorDiv);
            }
        } else if (existingError) {
            existingError.remove();
        }
    }

    static validateKeyForSubmission(keyValue) {
        for (let i = 0; i < keyValue.length; i++) {
            if (keyValue.charCodeAt(i) > 255) {
                return {
                    isValid: false,
                    error: 'Key can only contain ASCII characters (0-255)'
                };
            }
        }
        
        if (keyValue.length < 1) {
            return {
                isValid: false,
                error: 'Key cannot be empty'
            };
        }
        
        return { isValid: true };
    }

    static validateLSBBits(lsbValue) {
        const lsb = parseInt(lsbValue);
        if (isNaN(lsb) || lsb < 1 || lsb > 4) {
            return {
                isValid: false,
                error: 'LSB bits must be between 1 and 4'
            };
        }
        return { isValid: true };
    }

    static validateFile(fileInput, fileType = null) {
        if (!fileInput.files || fileInput.files.length === 0) {
            return {
                isValid: false,
                error: 'File is required'
            };
        }

        const file = fileInput.files[0];
        
        if (fileType === 'mp3' && !file.type.includes('audio')) {
            return {
                isValid: false,
                error: 'Please select a valid audio file'
            };
        }

        if (file.size > 100 * 1024 * 1024) {
            return {
                isValid: false,
                error: 'File size must be less than 100MB'
            };
        }

        return { isValid: true };
    }
}
