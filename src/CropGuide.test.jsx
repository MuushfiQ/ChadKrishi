// @vitest-environment jsdom
import React from 'react'
import { it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { CropGuide, ContainerSetup, ResearchCropCard } from './CropGuide'
import api from './api'
vi.mock('./api',()=>({default:{get:vi.fn(),put:vi.fn()},apiMessage:e=>e.message}))
beforeEach(()=>vi.clearAllMocks());afterEach(cleanup)
const setup={depthCm:null,volumeL:null,drainage:null,support:null,method:'seed'}
const crop={id:7,slug:'chili',name:'Chili',bengaliName:'মরিচ',status:'needs_setup',minSunHours:6,minDepthCm:null,minVolumeL:null,seedMonths:[9,10],transplantMonths:null,reasons:['Pot size is unknown.'],sourceIds:[],fieldSources:{}}
it('saves unknown dimensions as null, explicit false drainage, and chosen method',async()=>{
 api.put.mockResolvedValue({data:setup});render(<ContainerSetup initial={setup}/>);
 fireEvent.change(screen.getByLabelText('Drainage holes'),{target:{value:'false'}})
 fireEvent.change(screen.getByLabelText('Starting with'),{target:{value:'seedling'}})
 fireEvent.click(screen.getByText('Save container setup'))
 await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/garden-setup',expect.objectContaining({depthCm:null,volumeL:null,drainage:false,method:'seedling'})))
})
it('shows all catalog entries with unknown values and explains missing evidence',()=>{
 render(<MemoryRouter><CropGuide data={{setup,cropCatalog:[crop,{...crop,id:8,slug:'tomato',name:'Tomato',bengaliName:'টমেটো',status:'needs_evidence'}],sources:{}}}/></MemoryRouter>)
 expect(screen.getByText('মরিচ')).toBeTruthy();expect(screen.getByText('টমেটো')).toBeTruthy()
 expect(screen.getAllByText('Unknown').length).toBe(4)
 fireEvent.change(screen.getByLabelText('Show'),{target:{value:'needs_evidence'}})
 expect(screen.queryByText('মরিচ')).toBeNull();expect(screen.getByText('টমেটো')).toBeTruthy()
})
it('keeps failed saves visible and does not announce success',async()=>{
 api.put.mockRejectedValue(new Error('Database unavailable'));render(<ContainerSetup initial={setup}/>);
 fireEvent.click(screen.getByText('Save container setup'))
 expect(await screen.findByRole('alert')).toBeTruthy();expect(screen.queryByText('Container setup saved.')).toBeNull()
})

it('renders a local crop photo and falls back cleanly when missing',()=>{
 render(<MemoryRouter><ResearchCropCard crop={{...crop,imagePath:'/crops/chili.jpg'}}/></MemoryRouter>)
 const img=screen.getByRole('img',{name:'Chili'});expect(img.getAttribute('src')).toBe('/crops/chili.jpg')
 fireEvent.error(img);expect(screen.queryByRole('img')).toBeNull();expect(screen.getByText('মরিচ')).toBeTruthy()
})
it('does not load an arbitrary remote crop photo',()=>{
 render(<MemoryRouter><ResearchCropCard crop={{...crop,imagePath:'https://example.test/track.jpg'}}/></MemoryRouter>)
 expect(screen.queryByRole('img')).toBeNull()
})
