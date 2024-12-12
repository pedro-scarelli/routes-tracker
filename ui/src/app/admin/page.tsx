'use client'

import { useEffect, useRef } from 'react'
import { useMap } from '@/src/hooks/useMap'
import { socket } from '@/src/utils/socket-io'

export default function AdminPage() {
  const mapContainerRef = useRef<HTMLDivElement>(null)
  const map = useMap(mapContainerRef)

  useEffect(() => {
    if (!map) return

    // eslint-disable-next-line @typescript-eslint/no-unused-expressions
    socket.disconnected ? socket.connect() : socket.offAny()

    const eventKey = 'server:new-points:list'
    const handleNewPoints = async (data: { route_id: string; lat: number; lng: number }) => {
      const { route_id, lat, lng } = data
      if (!map.hasRoute(route_id)) {
        const response = await fetch(`http://localhost:3001/api/routes/${route_id}`)
        const route = await response.json()
        map.addRouteWithIcons({
          routeId: route_id,
          startMarkerOptions: {
            position: route.directions.routes[0].legs[0].start_location,
          },
          endMarkerOptions: {
            position: route.directions.routes[0].legs[0].end_location,
          },
          carMarkerOptions: {
            position: route.directions.routes[0].legs[0].start_location,
          },
        })
      }
      map.moveCar(route_id, { lat, lng })
    }

    socket.on(eventKey, handleNewPoints)
    return () => {
      socket.disconnect()
    }
  }, [map])

  return <div className="h-full w-full" ref={mapContainerRef}></div>
}
