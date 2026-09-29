/**
 * aplayer 类型声明（npm 官方无 @types/aplayer，此处按实际使用的 API 面声明）
 */
declare module 'aplayer' {
  export interface APlayerAudio {
    name: string
    artist?: string
    url: string
    cover?: string
    lrc?: string
    theme?: string
    type?: 'auto' | 'hls' | 'normal' | 'shaka'
  }

  export interface APlayerOptions {
    container: HTMLElement
    /** 固定悬浮模式（右下角） */
    fixed?: boolean
    /** 迷你模式 */
    mini?: boolean
    autoplay?: boolean
    theme?: string
    loop?: 'all' | 'one' | 'none'
    order?: 'list' | 'random'
    preload?: 'none' | 'metadata' | 'auto'
    volume?: number
    /** 是否互斥（同时只播放一个实例） */
    mutex?: boolean
    listFolded?: boolean
    listMaxHeight?: number
    /** 歌词类型：0 无 / 1 LRC 文件 / 2 LRC 文本 */
    lrcType?: number
    audio?: APlayerAudio[]
  }

  export default class APlayer {
    constructor(options: APlayerOptions)
    play(): void
    pause(): void
    seek(time: number): void
    toggle(): void
    on(event: string, handler: (...args: unknown[]) => void): void
    destroy(): void
  }
}
