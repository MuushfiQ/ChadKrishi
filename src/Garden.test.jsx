// @vitest-environment jsdom
import React from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor, cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { CarePanel, PlantsPage, FeedPage } from './Garden'
import api from './api'
vi.mock('./api',()=>({default:{get:vi.fn(),post:vi.fn(),put:vi.fn(),delete:vi.fn()},apiMessage:e=>e.message||'Request failed'}))
afterEach(cleanup)
beforeEach(()=>vi.clearAllMocks())
const wrapper=({children})=><MemoryRouter>{children}</MemoryRouter>
describe('garden workflows',()=>{
 it('preserves completed tasks when another task is checked',async()=>{
  api.get.mockResolvedValue({data:{tasks:[{id:1,plantName:'Brinjal',title:'Check soil',reason:'Observe first',completed:true},{id:2,plantName:'Brinjal',title:'Check leaves',reason:'Look for changes',completed:false}],guidance:'Observation prompts'}})
  api.put.mockResolvedValue({})
  render(<CarePanel compact/>,{wrapper})
  await screen.findByText('Check leaves');fireEvent.click(screen.getAllByRole('checkbox')[1])
  await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/tasks/2',{completed:true}))
  await waitFor(()=>expect(screen.getAllByRole('checkbox').every(c=>c.checked)).toBe(true))
 })
 it('sends a numeric crop ID and user-reported stage when adding a plant',async()=>{
  api.get.mockImplementation(url=>Promise.resolve({data:url==='/plants'?{plants:[]}:{crops:[{id:4,name:'Brinjal',verificationStatus:'source-backed'}]}}))
  api.post.mockResolvedValue({data:{id:8}})
  render(<PlantsPage/>,{wrapper});await screen.findByText('Your garden starts with one plant')
  fireEvent.click(screen.getByText('＋ Add a plant'))
  fireEvent.change(screen.getByLabelText('Crop'),{target:{value:'4'}})
  fireEvent.change(screen.getByLabelText('Plant nickname'),{target:{value:'Balcony friend'}})
  fireEvent.submit(screen.getByText('Save plant').closest('form'))
  await waitFor(()=>expect(api.post).toHaveBeenCalledWith('/plants',expect.objectContaining({cropId:4,nickname:'Balcony friend',growthStage:'seedling'})))
 })
 it('does not fabricate crops or posts after a failed feed request',async()=>{
  api.get.mockRejectedValue(new Error('Database unavailable'))
  render(<FeedPage/>,{wrapper});await screen.findByRole('alert')
  expect(screen.getByText('Database unavailable')).toBeTruthy()
  expect(screen.queryByText('Brinjal')).toBeNull()
 })
})
