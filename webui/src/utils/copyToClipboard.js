function fallbackCopyText(text) {
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.setAttribute('readonly', '')
    textArea.style.position = 'fixed'
    textArea.style.top = '-9999px'
    textArea.style.left = '-9999px'

    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()

    let copied = false
    try {
        copied = document.execCommand('copy')
    } finally {
        document.body.removeChild(textArea)
    }

    if (!copied) {
        throw new Error('copy failed')
    }
}

export async function copyToClipboard(text) {
    if (navigator.clipboard?.writeText) {
        try {
            await navigator.clipboard.writeText(text)
            return
        } catch {
            // fall through to fallback
        }
    }
    fallbackCopyText(text)
}
