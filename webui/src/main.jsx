import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App.jsx'
import { I18nProvider } from './i18n'
import './styles.css'

const basename = import.meta.env.MODE === 'production' ? '/admin' : '/'

ReactDOM.createRoot(document.getElementById('root')).render(
    <StrictMode>
        <I18nProvider>
            <BrowserRouter basename={basename}>
                <App />
            </BrowserRouter>
        </I18nProvider>
    </StrictMode>,
)
