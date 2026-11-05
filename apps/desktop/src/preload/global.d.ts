import type { ElectronAPI } from '@electron-toolkit/preload'
import type { AmiApi } from './index'

declare global {
  interface Window {
    electron: ElectronAPI
    api: AmiApi
  }
}

export {}
