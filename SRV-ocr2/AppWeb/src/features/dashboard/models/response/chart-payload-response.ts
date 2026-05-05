export interface ChartDataset<T = number> {
    label: string;
    data: T[];
    backgroundColor?: string | string[];
}

export interface ChartPayload {
    labels: string[];
    datasets: ChartDataset[];
    raw?: Record<string, number>;
}
