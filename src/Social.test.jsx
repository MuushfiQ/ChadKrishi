// @vitest-environment jsdom
import React from 'react'
import {beforeEach,afterEach,describe,it,expect,vi} from 'vitest'
import {render,screen,fireEvent,waitFor,cleanup} from '@testing-library/react'
import '@testing-library/jest-dom/vitest'
import {MemoryRouter, Routes, Route} from 'react-router-dom'
import {OrdersPage,MessagesPage,AdminPage} from './Social'
import {PostCard, FeedPage, PersonPage} from './Garden'
import {ProfilePhotoEditor} from './Avatar'
import api from './api'
vi.mock('./api',()=>({default:{get:vi.fn(),post:vi.fn(),put:vi.fn()},apiMessage:e=>e.message}))
afterEach(cleanup)
beforeEach(()=>{vi.clearAllMocks();api.put.mockResolvedValue({});vi.spyOn(window,'confirm').mockReturnValue(true)})
const order={id:1,title:'Rooftop plant',quantity:1,amountBdt:'250.00',isBuyer:false,buyerName:'Buyer',sellerName:'Seller',buyerId:2,sellerId:1,createdAt:'2026-10-22T01:00:00Z',status:'waiting_payment'}
describe('payment confirmation',()=>{
 it('seller cannot confirm before buyer declares payment',async()=>{api.get.mockResolvedValue({data:{orders:[order]}});render(<MemoryRouter><OrdersPage/></MemoryRouter>);expect(await screen.findByRole('button',{name:'Yes, payment received'})).toBeDisabled();expect(screen.queryByRole('button',{name:'I have paid'})).not.toBeInTheDocument()})
 it('buyer declares paid using method, without marking sold directly',async()=>{api.get.mockResolvedValue({data:{orders:[{...order,isBuyer:true}]}});render(<MemoryRouter><OrdersPage/></MemoryRouter>);fireEvent.change(await screen.findByLabelText('How did you pay?'),{target:{value:'bkash'}});fireEvent.click(screen.getByRole('button',{name:'I have paid'}));await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/orders/1',{action:'paid',method:'bkash',reference:''}));expect(screen.queryByRole('button',{name:'Yes, payment received'})).not.toBeInTheDocument()})
 it('seller confirms only after buyer paid',async()=>{api.get.mockResolvedValue({data:{orders:[{...order,status:'buyer_paid'}]}});render(<MemoryRouter><OrdersPage/></MemoryRouter>);fireEvent.click(await screen.findByRole('button',{name:'Yes, payment received'}));await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/orders/1',expect.objectContaining({action:'received'})))})
})
it('switching reaction sends one desired reaction and adjusts both counts',async()=>{const post={id:3,authorId:1,authorName:'Grower',type:'update',content:'Hello garden',createdAt:'2026-10-22',isOwner:true,reaction:'like',reactions:{like:1,love:0,angry:0,sad:0},commentCount:0};render(<MemoryRouter><PostCard post={post} onRefresh={()=>{}}/></MemoryRouter>);fireEvent.click(screen.getByRole('button',{name:'love reaction'}));await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/posts/3/reaction',{reaction:'love'}));await waitFor(()=>expect(screen.getByRole('button',{name:'love reaction'})).toHaveAttribute('aria-pressed','true'));expect(screen.getByRole('button',{name:'like reaction'})).toHaveAttribute('aria-pressed','false')})

it('private chat loads persisted text, marks displayed messages read, and sends new text',async()=>{
 let sent=false
 api.get.mockImplementation(async path=>path==='/conversations'?{data:{conversations:[{id:5,userId:2,name:'Other grower',lastMessage:'Hello',unread:1}],nextCursor:null}}:{data:{messages:sent?[{id:9,content:'New message',mine:true,createdAt:'2026-10-22T01:00:00Z'},{id:8,content:'Hello',mine:false,createdAt:'2026-10-22T00:00:00Z'}]:[{id:8,content:'Hello',mine:false,createdAt:'2026-10-22T00:00:00Z'}],nextCursor:null}})
 api.post.mockImplementation(async()=>{sent=true;return {data:{id:9}}})
 render(<MemoryRouter initialEntries={['/messages?thread=5']}><MessagesPage/></MemoryRouter>)
 await screen.findByRole('button',{name:'Send'})
 await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/conversations/5/read',{through:8}))
 fireEvent.change(screen.getByLabelText('Message'),{target:{value:'New message'}})
 fireEvent.click(screen.getByRole('button',{name:'Send'}))
 await waitFor(()=>expect(api.post).toHaveBeenCalledWith('/conversations/5/messages',{content:'New message'}))
 expect(await screen.findByText('New message',{selector:'p'})).toBeInTheDocument()
})
it('admin archive refreshes inventory and prevents repeated actions',async()=>{
 let archived=false
 api.get.mockImplementation(async()=>({data:{listings:[{id:4,title:'Sold seedling',sellerName:'Seller',archived}],nextCursor:null}}))
 api.put.mockImplementation(async()=>{archived=true;return {}})
 render(<MemoryRouter><AdminPage/></MemoryRouter>)
 fireEvent.click(await screen.findByRole('button',{name:'Archive listing'}))
 await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/admin/listings/4/archive'))
 expect(await screen.findByRole('button',{name:'Archived'})).toBeDisabled()
})

const listing={id:21,authorId:2,authorName:'Seller',area:'Dhaka',type:'sale',title:'Reserved seedling',content:'For your garden',createdAt:'2026-10-22',priceBdt:100,quantity:1,saleStatus:'reserved',isOwner:false,reactions:{},commentCount:0}
it('marketplace defaults to all listings and keeps reserved purchases visible without allowing a second buy',async()=>{
 api.get.mockResolvedValue({data:{posts:[listing],nextCursor:null}})
 render(<MemoryRouter><FeedPage market/></MemoryRouter>)
 expect(await screen.findByText('Reserved seedling')).toBeInTheDocument()
 expect(screen.getByRole('combobox',{name:'Sale status'})).toHaveValue('')
 expect(api.get).toHaveBeenCalledWith('/posts',expect.objectContaining({params:expect.objectContaining({type:'sale',status:''})}))
 expect(screen.getByRole('button',{name:'Reserved · awaiting payment confirmation'})).toBeDisabled()
 fireEvent.change(screen.getByRole('combobox',{name:'Sale status'}),{target:{value:'sold'}})
 await waitFor(()=>expect(api.get).toHaveBeenLastCalledWith('/posts',expect.objectContaining({params:expect.objectContaining({status:'sold'})})))
})
it('community author links to their profile and has no message action below the post',()=>{
 render(<MemoryRouter><PostCard post={{...listing,type:'update',authorPhoto:'/api/media/2-photo.png'}} onRefresh={()=>{}}/></MemoryRouter>)
 expect(screen.getByRole('link',{name:/Seller/})).toHaveAttribute('href','/people/2')
 expect(screen.getByRole('img',{name:"Seller's profile picture"})).toHaveAttribute('src','/api/media/2-photo.png')
 expect(screen.queryByRole('button',{name:/Message/})).not.toBeInTheDocument()
})
it('public profile shows the author photo, biography and previous posts and starts chat from its header',async()=>{
 api.get.mockImplementation(async path=>({data:path==='/people/2'?{person:{id:2,name:'Seller',area:'Dhaka',bio:'I grow rooftop plants',profilePhoto:'/api/media/2-photo.png',postCount:1,isSelf:false}}:{posts:[listing],nextCursor:null}}))
 api.post.mockResolvedValue({data:{id:7}})
 render(<MemoryRouter initialEntries={['/people/2']}><Routes><Route path='/people/:id' element={<PersonPage/>}/><Route path='/messages' element={<p>Chat opened</p>}/></Routes></MemoryRouter>)
 expect(await screen.findByText('I grow rooftop plants')).toBeInTheDocument()
 expect(await screen.findByText('Reserved seedling')).toBeInTheDocument()
 expect(api.get).toHaveBeenCalledWith('/posts',expect.objectContaining({params:expect.objectContaining({author:'2'})}))
 fireEvent.click(screen.getByRole('button',{name:/Message/}))
 await waitFor(()=>expect(api.post).toHaveBeenCalledWith('/conversations',{userId:2}))
 expect(await screen.findByText('Chat opened')).toBeInTheDocument()
})
it('own public profile offers editing instead of a self conversation',async()=>{
 api.get.mockImplementation(async path=>({data:path==='/people/2'?{person:{id:2,name:'Me',postCount:0,isSelf:true}}:{posts:[],nextCursor:null}}))
 render(<MemoryRouter initialEntries={['/people/2']}><Routes><Route path='/people/:id' element={<PersonPage/>}/></Routes></MemoryRouter>)
 expect(await screen.findByRole('link',{name:'Edit my profile'})).toHaveAttribute('href','/profile')
 expect(screen.queryByRole('button',{name:/Message/})).not.toBeInTheDocument()
})
it('profile picture upload persists its returned path and updates the signed-in user',async()=>{
 const onSaved=vi.fn(),user={id:1,name:'Grower'}
 api.post.mockResolvedValue({data:{path:'/api/media/1-new.png'}})
 api.put.mockResolvedValue({data:{user:{...user,profilePhoto:'/api/media/1-new.png'}}})
 render(<ProfilePhotoEditor user={user} onSaved={onSaved}/>)
 fireEvent.change(screen.getByLabelText('Profile picture'),{target:{files:[new File(['image'],'photo.png',{type:'image/png'})]}})
 await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/profile/photo',{photoPath:'/api/media/1-new.png'}))
 expect(api.post).toHaveBeenCalledWith('/uploads',expect.any(FormData),{headers:{'Content-Type':undefined}})
 await waitFor(()=>expect(onSaved).toHaveBeenCalledWith({...user,profilePhoto:'/api/media/1-new.png'}))
})
it('invalid profile picture is rejected before uploading',async()=>{
 render(<ProfilePhotoEditor user={{name:'Grower'}} onSaved={vi.fn()}/>)
 fireEvent.change(screen.getByLabelText('Profile picture'),{target:{files:[new File(['bad'],'bad.svg',{type:'image/svg+xml'})]}})
 expect(await screen.findByRole('alert')).toHaveTextContent('JPEG or PNG')
 expect(api.post).not.toHaveBeenCalled()
})
it('removing a profile picture persists the change and leaves garden setup independent',async()=>{
 const onSaved=vi.fn();api.put.mockResolvedValue({data:{user:{id:1,name:'Grower'}}})
 render(<ProfilePhotoEditor user={{id:1,name:'Grower',profilePhoto:'/api/media/1-old.png'}} onSaved={onSaved}/>)
 fireEvent.click(screen.getByRole('button',{name:'Remove profile picture'}))
 await waitFor(()=>expect(api.put).toHaveBeenCalledWith('/profile/photo',{photoPath:''}))
 await waitFor(()=>expect(onSaved).toHaveBeenCalledWith({id:1,name:'Grower'}))
})
it('failed picture save shows an error and does not claim success',async()=>{
 api.post.mockResolvedValue({data:{path:'/api/media/1-new.png'}})
 api.put.mockRejectedValueOnce(new Error('Database unavailable'))
 const onSaved=vi.fn();render(<ProfilePhotoEditor user={{id:1,name:'Grower'}} onSaved={onSaved}/>)
 fireEvent.change(screen.getByLabelText('Profile picture'),{target:{files:[new File(['image'],'photo.png',{type:'image/png'})]}})
 expect(await screen.findByRole('alert')).toHaveTextContent('Database unavailable')
 expect(onSaved).not.toHaveBeenCalled()
})
