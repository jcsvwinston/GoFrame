import { useState, useEffect } from 'react'
import { Search, Plus, Filter, ChevronLeft, ChevronRight } from 'lucide-react'

interface Product {
  id: number
  name: string
  price: number
  stock: number
  category: string
  image: string
  sku: string
}

function Products() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(100000)

  useEffect(() => {
    fetch(`/api/products?page=${page}`)
      .then(res => res.json())
      .then(data => {
        setProducts(data.products || [])
        setTotal(data.total || 100000)
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [page])

  const formatCurrency = (num: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(num)
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Products</h1>
          <p className="text-gray-500 mt-1">Manage your product catalog ({total.toLocaleString()} products)</p>
        </div>
        <button className="btn-primary flex items-center gap-2">
          <Plus size={20} />
          Add Product
        </button>
      </div>

      {/* Filters */}
      <div className="card flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
          <input
            type="text"
            placeholder="Search products..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>
        <button className="btn-secondary flex items-center gap-2">
          <Filter size={20} />
          Filters
        </button>
      </div>

      {/* Products Grid */}
      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {products.length > 0 ? (
              products.map((product) => (
                <div key={product.id} className="card p-4 hover:shadow-md transition-shadow">
                  <img
                    src={product.image || 'https://via.placeholder.com/300x200'}
                    alt={product.name}
                    className="w-full h-40 object-cover rounded-lg mb-4"
                  />
                  <h3 className="font-semibold text-gray-900 truncate">{product.name}</h3>
                  <p className="text-sm text-gray-500">{product.sku}</p>
                  <div className="flex items-center justify-between mt-3">
                    <span className="text-lg font-bold text-primary-600">{formatCurrency(product.price)}</span>
                    <span className={`text-sm ${product.stock > 10 ? 'text-green-600' : 'text-red-600'}`}>
                      {product.stock} in stock
                    </span>
                  </div>
                </div>
              ))
            ) : (
              // Mock data for demo
              Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="card p-4 hover:shadow-md transition-shadow">
                  <img
                    src={`https://picsum.photos/300/200?random=${i}`}
                    alt="Product"
                    className="w-full h-40 object-cover rounded-lg mb-4"
                  />
                  <h3 className="font-semibold text-gray-900 truncate">Premium Product {i + 1}</h3>
                  <p className="text-sm text-gray-500">SKU-{String(i + 1).padStart(6, '0')}</p>
                  <div className="flex items-center justify-between mt-3">
                    <span className="text-lg font-bold text-primary-600">{formatCurrency(99.99 + i * 10)}</span>
                    <span className="text-sm text-green-600">{100 - i * 10} in stock</span>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Pagination */}
          <div className="flex items-center justify-between">
            <p className="text-sm text-gray-500">
              Showing {((page - 1) * 20) + 1} to {Math.min(page * 20, total)} of {total.toLocaleString()} products
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="p-2 rounded-lg border border-gray-200 hover:bg-gray-50 disabled:opacity-50"
              >
                <ChevronLeft size={20} />
              </button>
              <span className="px-4 py-2 rounded-lg bg-primary-600 text-white font-medium">
                {page}
              </span>
              <button
                onClick={() => setPage(p => p + 1)}
                className="p-2 rounded-lg border border-gray-200 hover:bg-gray-50"
              >
                <ChevronRight size={20} />
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}

export default Products
