import { Injectable } from '@nestjs/common';
import { PrismaService } from 'src/prisma/prisma.service';
import { UpsertPointDto } from './dto/upsert-point.dto';

@Injectable()
export class RoutesDriverService {
  constructor(private prismaService: PrismaService) {}
  processRoute(upsertPointDto: UpsertPointDto) {
    const { route_id, lat, lng } = upsertPointDto;
    return this.prismaService.routeDriver.upsert({
      include: {
        route: true,
      },
      where: { route_id: route_id },
      create: {
        route_id,
        points: {
          set: {
            location: {
              lat,
              lng,
            },
          },
        },
      },
      update: {
        points: {
          push: {
            location: {
              lat,
              lng,
            },
          },
        },
      },
    });
  }
}
