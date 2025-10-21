'use client'
import dynamic from 'next/dynamic'

const MeetNote = dynamic(() => import('./client-page'), { 
  ssr: false,
  loading: () => <div className="min-h-screen flex items-center justify-center">
      Loading Noted...
    </div>
})

export default function Page() {
  return <MeetNote />
}