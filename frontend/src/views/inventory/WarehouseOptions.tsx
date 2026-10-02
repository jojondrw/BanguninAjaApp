import type { Warehouse } from '../../models/inventory'

export function WarehouseOptions({ warehouses, placeholder }: { warehouses: Warehouse[]; placeholder: string }) {
  return (
    <>
      <option value="">{placeholder}</option>
      {warehouses.map((warehouse) => (
        <option key={warehouse.id} value={warehouse.id}>
          {warehouse.name} ({warehouse.code})
        </option>
      ))}
    </>
  )
}
