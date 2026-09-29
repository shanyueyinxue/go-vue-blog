import { describe, expect, it } from 'vitest'
import { parseCustomCssVars, stripCssComments } from '@/stores/site'

describe('stripCssComments', () => {
  it('移除值后面的块注释', () => {
    expect(stripCssComments('--blog-primary: #66afef; /* 主色 */')).toBe('--blog-primary: #66afef; ')
  })

  it('移除分号前的块注释', () => {
    expect(stripCssComments('--x: #fff /* c */;')).toBe('--x: #fff ;')
  })

  it('移除行注释', () => {
    expect(stripCssComments('--font-size: 16px // 字号')).toBe('--font-size: 16px ')
  })

  it('保留引号内的 //', () => {
    expect(stripCssComments('--content: "http://a.com";')).toBe('--content: "http://a.com";')
  })

  it('保留 url() 内的 //', () => {
    expect(stripCssComments('--bg: url(//cdn.example.com/a.png);')).toBe('--bg: url(//cdn.example.com/a.png);')
  })
})

describe('parseCustomCssVars', () => {
  it('解析基础变量', () => {
    expect(parseCustomCssVars('--blog-primary: #66afef;')).toEqual([['--blog-primary', '#66afef']])
  })

  it('值后面的注释不会混入变量值', () => {
    expect(parseCustomCssVars('--blog-primary: #66afef; /* 主色 */')).toEqual([
      ['--blog-primary', '#66afef'],
    ])
    expect(parseCustomCssVars('--font-size: 16px // 字号')).toEqual([['--font-size', '16px']])
  })

  it('忽略空行与纯注释行', () => {
    expect(parseCustomCssVars('\n// 注释\n/* 块 */\n--x: 1;')).toEqual([['--x', '1']])
  })
})
