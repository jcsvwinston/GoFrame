import { useState, useEffect } from 'react'
import { Search, Mail, Phone, MapPin } from 'lucide-react'

interface Customer {
  id: number
  name: string
  email: string
  phone: string
  address: string
  orders: number
  totalSpent: number
}

function Customers() {
  const [customers, setCustomers] = useState<Customer[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')

  useEffect(() => {
    // Mock data
    const mockCustomers: Customer[] = Array.from({ length: 20 }, (_, i) => ({
      id: i + 1,
      name: `Customer ${i + 1}`,
      email: `customer${i + 1}@example.com`,
      phone: `(555) ${String(Math.floor(Math.random() * 1000)).padStart(3, '0')}-${String(Math.floor(Math.random() * 10000)).padStart(4, '0')}`,
      address: `${Math.floor(Math.random() * 9999) + 1} Main St, City ${i + 1}`,
      orders: Math.floor(Math.random() * 20) + 1,
      totalSpent: Math.random() * 5000 + 100,
    }))
    setCustomers(mockCustomers)
    setLoading(false)
  }, [])

  const formatCurrency = (num: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(num)
  }

  const filteredCustomers = customers.filter(customer =>
    customer.name.toLowerCase().includes(search.toLowerCase()) ||
    customer.email.toLowerCase().includes(search.toLowerCase())
  )

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Customers</h1>
        <p className="text-gray-500 mt-1">Manage your customer base (50,000+ total)</p>
      </div>

      {/* Search */}
      <div className="card">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
          <input
            type="text"
            placeholder="Search customers by name or email..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>
      </div>

      {/* Customers Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredCustomers.map((customer) => (
          <div key={customer.id} className="card">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 rounded-full bg-primary-100 flex items-center justify-center text-primary-600 font-bold text-lg">
                {customer.name.charAt(8)}
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold text-gray-900 truncate">{customer.name}</h3>
                <div className="flex items-center gap-2 mt-1 text-sm text-gray-500">
                  <Mail size={14} />
                  <span className="truncate">{customer.email}</span>
                </div>
                <div className="flex items-center gap-2 mt-1 text-sm text-gray-500">
                  <Phone size={14} />
                  <span>{customer.phone}</span>
                </div>
                <div className="flex items-center gap-2 mt-1 text-sm text-gray-500">
                  <MapPin size={14} />
                  <span className="truncate">{customer.address}</span>
                </div>
              </div>
            </div>
            <div className="mt-4 pt-4 border-t border-gray-200 flex items-center justify-between">
              <div>
                <p className="text-xs text-gray-500">Orders</p>
                <p className="font-semibold text-gray-900">{customer.orders}</p>
              </div>
              <div className="text-right">
                <p className="text-xs text-gray-500">Total Spent</p>
                <p className="font-semibold text-primary-600">{formatCurrency(customer.totalSpent)}</p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default Customers
