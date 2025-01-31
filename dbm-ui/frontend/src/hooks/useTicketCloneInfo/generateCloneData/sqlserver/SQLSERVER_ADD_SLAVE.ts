import TicketModel, { type Sqlserver } from '@services/model/ticket/ticket';

export default async (ticketDetail: TicketModel<Sqlserver.AddSlave>) =>
  ticketDetail.details.infos.map((item) => ({
    cluster_ids: item.cluster_ids,
    clusters: item.cluster_ids.map((item) => ticketDetail.details.clusters[item]),
    new_slave_host: item.new_slave_host,
  }));
